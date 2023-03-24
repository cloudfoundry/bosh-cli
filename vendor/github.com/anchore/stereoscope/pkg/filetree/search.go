package filetree

import (
	"fmt"
	"path"
	"sort"

	"github.com/anchore/stereoscope/internal/log"
	"github.com/anchore/stereoscope/pkg/filetree/filenode"
	"github.com/anchore/stereoscope/pkg/tree/node"

	"github.com/anchore/stereoscope/pkg/file"
	"github.com/bmatcuk/doublestar/v4"
)

// Searcher is a facade for searching a file tree with optional indexing support.
type Searcher interface {
	SearchByPath(path string, options ...LinkResolutionOption) (*file.Resolution, error)
	SearchByGlob(patterns string, options ...LinkResolutionOption) ([]file.Resolution, error)
	SearchByMIMEType(mimeTypes ...string) ([]file.Resolution, error)
}

type searchContext struct {
	tree  *FileTree   // this is the tree which all index search results are filtered against
	index IndexReader // this index is relative to one or more trees, not just necessarily one

	// the following enables correct link resolution when searching via the index
	linkBackwardRefs map[node.ID]node.IDSet // {link-destination-node-id: str([link-node-id, ...])}
}

func NewSearchContext(tree Reader, index IndexReader) Searcher {
	c := &searchContext{
		tree:             tree.(*FileTree),
		index:            index,
		linkBackwardRefs: make(map[node.ID]node.IDSet),
	}

	if err := c.buildLinkResolutionIndex(); err != nil {
		log.WithFields("error", err).Warn("unable to build link resolution index for filetree search context")
	}

	return c
}

func (sc *searchContext) buildLinkResolutionIndex() error {
	entries, err := sc.index.GetByFileType(file.TypeSymLink, file.TypeHardLink)
	if err != nil {
		return err
	}

	// filter the results relative to the tree
	nodes, err := sc.fileNodesInTree(entries)
	if err != nil {
		return err
	}

	// note: the remaining references are all links that exist in the tree

	for _, fn := range nodes {
		destinationFna, err := sc.tree.file(fn.RenderLinkDestination())
		if err != nil {
			return fmt.Errorf("unable to get node for path=%q: %w", fn.RealPath, err)
		}

		if !destinationFna.HasFileNode() {
			// we were unable to resolve the link destination, this could be due to the fact that the destination simply
			continue
		}

		linkID := fn.ID()
		destinationID := destinationFna.FileNode.ID()

		// add backward reference...
		if _, ok := sc.linkBackwardRefs[destinationID]; !ok {
			sc.linkBackwardRefs[destinationID] = node.NewIDSet()
		}
		sc.linkBackwardRefs[destinationID].Add(linkID)
	}

	return nil
}

func (sc searchContext) SearchByPath(path string, options ...LinkResolutionOption) (*file.Resolution, error) {
	// TODO: one day this could leverage indexes outside of the tree, but today this is not implemented
	log.WithFields("path", path).Trace("searching filetree by path")

	options = append(options, FollowBasenameLinks)
	_, ref, err := sc.tree.File(file.Path(path), options...)
	return ref, err
}

func (sc searchContext) SearchByMIMEType(mimeTypes ...string) ([]file.Resolution, error) {
	log.WithFields("types", mimeTypes).Trace("searching filetree by MIME types")

	var fileEntries []IndexEntry

	for _, mType := range mimeTypes {
		entries, err := sc.index.GetByMIMEType(mType)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch file references by MIME type (%q): %w", mType, err)
		}
		fileEntries = append(fileEntries, entries...)
	}

	refs, err := sc.referencesInTree(fileEntries)
	if err != nil {
		return nil, err
	}

	sort.Sort(file.Resolutions(refs))

	return refs, nil
}

// add case for status.d/* like things that hook up directly into filetree.ListPaths()

func (sc searchContext) SearchByGlob(pattern string, options ...LinkResolutionOption) ([]file.Resolution, error) {
	log.WithFields("glob", pattern).Trace("searching filetree by glob")

	if sc.index == nil {
		options = append(options, FollowBasenameLinks)
		refs, err := sc.tree.FilesByGlob(pattern, options...)
		if err != nil {
			return nil, fmt.Errorf("unable to search by glob=%q: %w", pattern, err)
		}
		sort.Sort(file.Resolutions(refs))
		return refs, nil
	}

	var allRefs []file.Resolution
	for _, request := range parseGlob(pattern) {
		refs, err := sc.searchByRequest(request, options...)
		if err != nil {
			return nil, fmt.Errorf("unable to search by glob=%q: %w", pattern, err)
		}
		allRefs = append(allRefs, refs...)
	}

	sort.Sort(file.Resolutions(allRefs))

	return allRefs, nil
}

func (sc searchContext) searchByRequest(request searchRequest, options ...LinkResolutionOption) ([]file.Resolution, error) {
	switch request.searchBasis {
	case searchByFullPath:
		options = append(options, FollowBasenameLinks)
		ref, err := sc.SearchByPath(request.value, options...)
		if err != nil {
			return nil, err
		}
		if ref == nil {
			return nil, nil
		}
		return []file.Resolution{*ref}, nil
	case searchByBasename:
		indexes, err := sc.index.GetByBasename(request.value)
		if err != nil {
			return nil, fmt.Errorf("unable to search by basename=%q: %w", request.value, err)
		}
		refs, err := sc.referencesWithRequirement(request.requirement, indexes)
		if err != nil {
			return nil, err
		}
		return refs, nil
	case searchByBasenameGlob:
		indexes, err := sc.index.GetByBasenameGlob(request.value)
		if err != nil {
			return nil, fmt.Errorf("unable to search by basename-glob=%q: %w", request.value, err)
		}
		refs, err := sc.referencesWithRequirement(request.requirement, indexes)
		if err != nil {
			return nil, err
		}
		return refs, nil
	case searchByExtension:
		indexes, err := sc.index.GetByExtension(request.value)
		if err != nil {
			return nil, fmt.Errorf("unable to search by extension=%q: %w", request.value, err)
		}
		refs, err := sc.referencesWithRequirement(request.requirement, indexes)
		if err != nil {
			return nil, err
		}
		return refs, nil
	case searchBySubDirectory:
		return sc.searchByParentBasename(request)

	case searchByGlob:
		log.WithFields("glob", request.value).Trace("glob provided is an expensive search, consider using a more specific indexed search")

		options = append(options, FollowBasenameLinks)
		return sc.tree.FilesByGlob(request.value, options...)
	}

	return nil, fmt.Errorf("invalid search request: %+v", request.searchBasis)
}

func (sc searchContext) searchByParentBasename(request searchRequest) ([]file.Resolution, error) {
	indexes, err := sc.index.GetByBasename(request.value)
	if err != nil {
		return nil, fmt.Errorf("unable to search by extension=%q: %w", request.value, err)
	}
	refs, err := sc.referencesWithRequirement(request.requirement, indexes)
	if err != nil {
		return nil, err
	}

	var results []file.Resolution
	for _, ref := range refs {
		paths, err := sc.tree.ListPaths(ref.RequestPath)
		if err != nil {
			// this may not be a directory, that's alright, just continue...
			continue
		}
		for _, p := range paths {
			_, nestedRef, err := sc.tree.File(p, FollowBasenameLinks)
			if err != nil {
				return nil, fmt.Errorf("unable to fetch file reference from parent path %q for path=%q: %w", ref.RequestPath, p, err)
			}
			if !nestedRef.HasReference() {
				continue
			}
			// note: the requirement was written for the parent, so we need to consider the new
			// child path by adding /* to match all children
			matches, err := matchesRequirement(*nestedRef, request.requirement+"/*")
			if err != nil {
				return nil, err
			}
			if matches {
				results = append(results, *nestedRef)
			}
		}
	}

	return results, nil
}

func (sc searchContext) referencesWithRequirement(requirement string, entries []IndexEntry) ([]file.Resolution, error) {
	refs, err := sc.referencesInTree(entries)
	if err != nil {
		return nil, err
	}

	if requirement == "" {
		return refs, nil
	}

	var results []file.Resolution
	for _, ref := range refs {
		matches, err := matchesRequirement(ref, requirement)
		if err != nil {
			return nil, err
		}
		if matches {
			results = append(results, ref)
		}
	}

	return results, nil
}

func matchesRequirement(ref file.Resolution, requirement string) (bool, error) {
	allRefPaths := ref.AllRequestPaths()
	for _, p := range allRefPaths {
		matched, err := doublestar.Match(requirement, string(p))
		if err != nil {
			return false, fmt.Errorf("unable to match glob pattern=%q to path=%q: %w", requirement, p, err)
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

type cacheRequest struct {
	RealPath file.Path
}

type cacheResult struct {
	Paths file.PathSet
	Error error
}

func (sc searchContext) allPathsToNode(fn *filenode.FileNode) ([]file.Path, error) {
	if fn == nil {
		return nil, nil
	}

	observedPaths := file.NewPathSet()

	cache := map[cacheRequest]cacheResult{}

	paths, err := sc.pathsToNode(fn, observedPaths, cache)
	if err != nil {
		return nil, err
	}

	pathsList := paths.List()
	sort.Sort(file.Paths(pathsList))

	// TODO: filter to only paths that exist in the tree

	return pathsList, nil
}

func (sc searchContext) pathsToNode(fn *filenode.FileNode, observedPaths file.PathSet, cache map[cacheRequest]cacheResult) (file.PathSet, error) {
	req := cacheRequest{
		RealPath: fn.RealPath,
	}

	if result, ok := cache[req]; ok {
		return result.Paths, result.Error
	}

	paths, err := sc._pathsToNode(fn, observedPaths, cache)

	cache[req] = cacheResult{
		Paths: paths,
		Error: err,
	}

	return paths, err
}

// nolint: funlen
func (sc searchContext) _pathsToNode(fn *filenode.FileNode, observedPaths file.PathSet, cache map[cacheRequest]cacheResult) (file.PathSet, error) {
	if fn == nil {
		return nil, nil
	}

	paths := file.NewPathSet()
	paths.Add(fn.RealPath)

	if observedPaths != nil {
		if observedPaths.Contains(fn.RealPath) {
			// we've already observed this path, so we can stop here
			return nil, nil
		}
		observedPaths.Add(fn.RealPath)
	}

	nodeID := fn.ID()

	addPath := func(suffix string, ps ...file.Path) {
		for _, p := range ps {
			if suffix != "" {
				p = file.Path(path.Join(string(p), suffix))
			}
			paths.Add(p)
		}
	}

	// add all paths to the node that are linked to it
	for _, linkSrcID := range sc.linkBackwardRefs[nodeID].List() {
		pfn := sc.tree.tree.Node(linkSrcID)
		if pfn == nil {
			log.WithFields("id", nodeID, "parent", linkSrcID).Warn("found non-existent parent link")
			continue
		}
		linkSrcPaths, err := sc.pathsToNode(pfn.(*filenode.FileNode), observedPaths, cache)
		if err != nil {
			return nil, err
		}

		addPath("", linkSrcPaths.List()...)
	}

	// crawl up the tree to find all paths to our parent and repeat
	for _, p := range paths.List() {
		nextNestedSuffix := p.Basename()
		allParentPaths := p.ConstituentPaths()
		sort.Sort(sort.Reverse(file.Paths(allParentPaths)))

		for _, pp := range allParentPaths {
			if pp == "/" {
				break
			}

			nestedSuffix := nextNestedSuffix
			nextNestedSuffix = path.Join(pp.Basename(), nestedSuffix)

			pna, err := sc.tree.node(pp, linkResolutionStrategy{
				FollowAncestorLinks: true,
				FollowBasenameLinks: false,
			})
			if err != nil {
				return nil, fmt.Errorf("unable to get parent node for path=%q: %w", pp, err)
			}

			if !pna.HasFileNode() {
				continue
			}

			parentLinkPaths, err := sc.pathsToNode(pna.FileNode, observedPaths, cache)
			if err != nil {
				return nil, err
			}
			addPath(nestedSuffix, parentLinkPaths.List()...)
		}
	}
	observedPaths.Remove(fn.RealPath)

	return paths, nil
}

func (sc searchContext) fileNodesInTree(fileEntries []IndexEntry) ([]*filenode.FileNode, error) {
	var nodes []*filenode.FileNode
allFileEntries:
	for _, entry := range fileEntries {
		// note: it is important that we don't enable any basename link resolution
		na, err := sc.tree.file(entry.Reference.RealPath)
		if err != nil {
			return nil, fmt.Errorf("unable to get ref for path=%q: %w", entry.Reference.RealPath, err)
		}

		if !na.HasFileNode() {
			continue
		}

		// only check the resolved node matches the index entries reference, not via link resolutions...
		if na.FileNode.Reference != nil && na.FileNode.Reference.ID() == entry.Reference.ID() {
			nodes = append(nodes, na.FileNode)
			continue allFileEntries
		}

		// we did not find a matching file ID in the tree, so drop this entry
	}
	return nodes, nil
}

// referencesInTree does two things relative to the index entries given:
// 1) it expands the index entries to include all possible access paths to the file node (by considering all possible link resolutions)
// 2) it filters the index entries to only include those that exist in the tree
func (sc searchContext) referencesInTree(fileEntries []IndexEntry) ([]file.Resolution, error) {
	var refs []file.Resolution

	for _, entry := range fileEntries {
		na, err := sc.tree.file(entry.Reference.RealPath, FollowBasenameLinks)
		if err != nil {
			return nil, fmt.Errorf("unable to get ref for path=%q: %w", entry.Reference.RealPath, err)
		}

		// this filters out any index entries that do not exist in the tree
		if !na.HasFileNode() {
			continue
		}

		// expand the index results with more possible access paths from the link resolution cache
		var expandedRefs []file.Resolution
		allPathsToNode, err := sc.allPathsToNode(na.FileNode)
		if err != nil {
			return nil, fmt.Errorf("unable to get all paths to node for path=%q: %w", entry.Reference.RealPath, err)
		}
		for _, p := range allPathsToNode {
			_, ref, err := sc.tree.File(p, FollowBasenameLinks)
			if err != nil {
				return nil, fmt.Errorf("unable to get ref for path=%q: %w", p, err)
			}
			if !ref.HasReference() {
				continue
			}
			expandedRefs = append(expandedRefs, *ref)
		}

		for _, ref := range expandedRefs {
			for _, accessRef := range ref.References() {
				if accessRef.ID() == entry.Reference.ID() {
					// we know this entry exists in the tree, keep track of the reference for this file
					refs = append(refs, ref)
				}
			}
		}
	}
	return refs, nil
}
