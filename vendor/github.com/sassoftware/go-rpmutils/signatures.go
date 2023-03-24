/*
 * Copyright (c) SAS Institute Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rpmutils

import (
	"bytes"
	"crypto"
	"crypto/md5"
	"errors"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/sassoftware/go-rpmutils/fileutil"
	"golang.org/x/crypto/openpgp/packet"
)

// SignatureOptions describes additional configuration for SignRpm methods
type SignatureOptions struct {
	// Hash algorithm for the signature. If not set, defaults to SHA-256
	Hash crypto.Hash
	// CreationTime for the signature. If not set, defaults to the current time
	CreationTime time.Time
}

func (opts *SignatureOptions) hash() crypto.Hash {
	if opts != nil {
		return opts.Hash
	}
	return crypto.SHA256
}

func (opts *SignatureOptions) creationTime() time.Time {
	if opts != nil {
		return opts.CreationTime
	}
	return time.Now()
}

func makeSignature(stream io.Reader, key *packet.PrivateKey, opts *SignatureOptions) ([]byte, error) {
	hash := opts.hash()
	sig := &packet.Signature{
		SigType:      packet.SigTypeBinary,
		CreationTime: opts.creationTime(),
		PubKeyAlgo:   key.PublicKey.PubKeyAlgo,
		Hash:         hash,
		IssuerKeyId:  &key.KeyId,
	}
	h := hash.New()
	_, err := io.Copy(h, stream)
	if err != nil {
		return nil, err
	}
	err = sig.Sign(h, key, nil)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = sig.Serialize(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func insertSignature(sigHeader *rpmHeader, tag int, value []byte) {
	sigHeader.entries[tag] = entry{
		dataType: RPM_BIN_TYPE,
		count:    int32(len(value)),
		contents: value,
	}
}

func insertSignatures(sigHeader *rpmHeader, sigPgp, sigRsa []byte) {
	insertSignature(sigHeader, SIG_PGP-_SIGHEADER_TAG_BASE, sigPgp)
	insertSignature(sigHeader, SIG_RSA, sigRsa)
	delete(sigHeader.entries, SIG_GPG-_SIGHEADER_TAG_BASE)
	delete(sigHeader.entries, SIG_DSA)
}

func getSha1(sigHeader *rpmHeader) string {
	vals, err := sigHeader.GetStrings(SIG_SHA1)
	if err != nil {
		return ""
	}
	return vals[0]
}

func checkMd5(sigHeader *rpmHeader, h hash.Hash) bool {
	sigmd5, err := sigHeader.GetBytes(SIG_MD5 - _SIGHEADER_TAG_BASE)
	if err != nil {
		return true
	}
	return bytes.Equal(sigmd5, h.Sum(nil))
}

// SignRpmStream reads an RPM and signs it, returning the set of headers updated with the new signature.
func SignRpmStream(stream io.Reader, key *packet.PrivateKey, opts *SignatureOptions) (header *RpmHeader, err error) {
	lead, sigHeader, err := readSignatureHeader(stream)
	if err != nil {
		return
	}
	// parse the general header and also tee it into a buffer
	genHeaderBuf := new(bytes.Buffer)
	headerTee := io.TeeReader(stream, genHeaderBuf)
	genHeader, err := readHeader(headerTee, getSha1(sigHeader), sigHeader.isSource, false)
	if err != nil {
		return
	}
	genHeaderBlob := genHeaderBuf.Bytes()
	// chain the buffered general header to the rest of the payload, and digest the whole lot of it
	genHeaderAndPayload := io.MultiReader(bytes.NewReader(genHeaderBlob), stream)
	payloadDigest := md5.New()
	payloadTee := io.TeeReader(genHeaderAndPayload, payloadDigest)
	sigPgp, err := makeSignature(payloadTee, key, opts)
	if err != nil {
		return
	}
	if !checkMd5(sigHeader, payloadDigest) {
		return nil, errors.New("md5 digest mismatch")
	}
	sigRsa, err := makeSignature(bytes.NewReader(genHeaderBlob), key, opts)
	if err != nil {
		return
	}
	insertSignatures(sigHeader, sigPgp, sigRsa)
	return &RpmHeader{
		lead:      lead,
		sigHeader: sigHeader,
		genHeader: genHeader,
		isSource:  sigHeader.isSource,
	}, nil
}

func canOverwrite(ininfo, outinfo os.FileInfo) bool {
	if !outinfo.Mode().IsRegular() {
		return false
	}
	if !os.SameFile(ininfo, outinfo) {
		return false
	}
	if fileutil.HasLinks(outinfo) {
		return false
	}
	return true
}

// SignRpmFile signs infile and writes it to outpath, which may be the same file
func SignRpmFile(infile *os.File, outpath string, key *packet.PrivateKey, opts *SignatureOptions) (header *RpmHeader, err error) {
	header, err = SignRpmStream(infile, key, opts)
	if err != nil {
		return
	}
	return header, rewriteRpm(infile, outpath, header)
}

// RewriteWithSignatures inserts raw signatures into a RPM header.
//
// DEPRECATED: To perform a detached signature, use SignRpmStream and call
// DumpSignatureHeader to export the result.
func RewriteWithSignatures(infile *os.File, outpath string, sigPgp, sigRsa []byte) (*RpmHeader, error) {
	header, err := ReadHeader(infile)
	if err != nil {
		return nil, err
	}
	insertSignatures(header.sigHeader, sigPgp, sigRsa)
	err = rewriteRpm(infile, outpath, header)
	if err != nil {
		return nil, err
	}
	return header, nil
}

func rewriteRpm(infile *os.File, outpath string, header *RpmHeader) error {
	delete(header.sigHeader.entries, SIG_RESERVEDSPACE-_SIGHEADER_TAG_BASE)
	ininfo, err := infile.Stat()
	if err != nil {
		return err
	}
	var outstream io.Writer
	if outpath == "-" {
		outstream = os.Stdout
	} else {
		outinfo, err := os.Lstat(outpath)
		if err == nil && canOverwrite(ininfo, outinfo) {
			ok, err := writeInPlace(outpath, header)
			if err != nil || ok {
				return err
			}
			// in-place didn't work; fallback to rewrite
		} else if err == nil && !outinfo.Mode().IsRegular() {
			// pipe or something else. open for writing.
			outfile, err := os.Create(outpath)
			if err != nil {
				return err
			}
			defer outfile.Close()
			outstream = outfile
		}
		if outstream == nil {
			// write-rename
			tempfile, err := ioutil.TempFile(path.Dir(outpath), path.Base(outpath))
			if err != nil {
				return err
			}
			defer func() {
				if err != nil {
					os.Remove(tempfile.Name())
				} else {
					_ = tempfile.Chmod(0644)
					if err = tempfile.Close(); err != nil {
						return
					}
					err = os.Rename(tempfile.Name(), outpath)
				}
			}()
			outstream = tempfile
		}
	}
	return writeRpm(infile, outstream, header.sigHeader)
}

func writeInPlace(path string, header *RpmHeader) (ok bool, err error) {
	blob, err := header.DumpSignatureHeader(true)
	if err != nil {
		return false, err
	}
	orig := header.OriginalSignatureHeaderSize()
	if orig != len(blob) {
		// size changed; can't rewrite in place
		return false, nil
	}
	outfile, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return
	}
	defer outfile.Close()
	n, err := outfile.Write(blob)
	if err != nil {
		return false, err
	} else if n != len(blob) {
		return false, io.ErrShortWrite
	}
	return true, nil
}

func writeRpm(infile io.ReadSeeker, outstream io.Writer, sigHeader *rpmHeader) error {
	if _, err := infile.Seek(0, 0); err != nil {
		return err
	}
	lead, err := readExact(infile, 96)
	if err != nil {
		return err
	}
	if _, err = outstream.Write(lead); err != nil {
		return err
	}
	if err = sigHeader.WriteTo(outstream, RPMTAG_HEADERSIGNATURES); err != nil {
		return err
	}
	if _, err := infile.Seek(int64(len(lead)+sigHeader.origSize), 0); err != nil {
		return err
	}
	_, err = io.Copy(outstream, infile)
	return err
}
