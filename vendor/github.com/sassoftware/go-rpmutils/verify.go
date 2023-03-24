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
	"fmt"
	"hash"
	"io"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

var headerSigTags = []int{SIG_RSA, SIG_DSA}
var payloadSigTags = []int{
	SIG_PGP - _SIGHEADER_TAG_BASE,
	SIG_GPG - _SIGHEADER_TAG_BASE,
}

// Signature describes a PGP signature found within a RPM while verifying it.
type Signature struct {
	// Signer is the PGP identity that created the signature. It may be nil if
	// the public key is not available at verification time, but KeyId will
	// always be set.
	Signer *openpgp.Entity
	// Hash is the algorithm used to digest the signature contents
	Hash crypto.Hash
	// CreationTime is when the signature was created
	CreationTime time.Time
	// HeaderOnly is true for signatures that only cover the general RPM header,
	// and false for signatures that cover the general header plus the payload
	HeaderOnly bool
	// KeyId is the PGP key that created the signature.
	KeyId uint64

	packet packet.Packet
	hash   hash.Hash
}

// Verify the PGP signature over a RPM file. knownKeys should enumerate public
// keys to check against, otherwise the signature validity cannot be verified.
// If knownKeys is nil then digests will be checked but only the raw key ID will
// be available.
func Verify(stream io.Reader, knownKeys openpgp.EntityList) (header *RpmHeader, sigs []*Signature, err error) {
	lead, sigHeader, err := readSignatureHeader(stream)
	if err != nil {
		return nil, nil, err
	}
	payloadDigest := md5.New()
	sigs, headerMulti, payloadMulti, err := setupDigesters(sigHeader, payloadDigest)
	if err != nil {
		return nil, nil, err
	}
	// parse the general header and also tee it into a buffer
	genHeaderBuf := new(bytes.Buffer)
	headerTee := io.TeeReader(stream, genHeaderBuf)
	genHeader, err := readHeader(headerTee, getSha1(sigHeader), sigHeader.isSource, false)
	if err != nil {
		return nil, nil, err
	}
	genHeaderBlob := genHeaderBuf.Bytes()
	if _, err := headerMulti.Write(genHeaderBlob); err != nil {
		return nil, nil, err
	}
	// chain the buffered general header to the rest of the payload, and digest the whole lot of it
	genHeaderAndPayload := io.MultiReader(bytes.NewReader(genHeaderBlob), stream)
	if _, err := io.Copy(payloadMulti, genHeaderAndPayload); err != nil {
		return nil, nil, err
	}
	if !checkMd5(sigHeader, payloadDigest) {
		return nil, nil, errors.New("md5 digest mismatch")
	}
	if knownKeys != nil {
		for _, sig := range sigs {
			if err := checkSig(sig, knownKeys); err != nil {
				return nil, nil, err
			}
		}
	}
	hdr := &RpmHeader{
		lead:      lead,
		sigHeader: sigHeader,
		genHeader: genHeader,
		isSource:  sigHeader.isSource,
	}
	return hdr, sigs, nil
}

func setupDigester(sigHeader *rpmHeader, tag int) (*Signature, error) {
	blob, err := sigHeader.GetBytes(tag)
	if _, ok := err.(NoSuchTagError); ok {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	packetReader := packet.NewReader(bytes.NewReader(blob))
	genpkt, err := packetReader.Next()
	if err != nil {
		return nil, err
	}
	var sig *Signature
	switch pkt := genpkt.(type) {
	case *packet.SignatureV3:
		sig = &Signature{
			Hash:         pkt.Hash,
			CreationTime: pkt.CreationTime,
			KeyId:        pkt.IssuerKeyId,
		}
	case *packet.Signature:
		if pkt.IssuerKeyId == nil {
			return nil, errors.New("Missing keyId in signature")
		}
		sig = &Signature{
			Hash:         pkt.Hash,
			CreationTime: pkt.CreationTime,
			KeyId:        *pkt.IssuerKeyId,
		}
	default:
		return nil, fmt.Errorf("tag %d does not contain a PGP signature", tag)
	}
	_, err = packetReader.Next()
	if err != io.EOF {
		return nil, fmt.Errorf("trailing garbage after signature in tag %d", tag)
	}
	sig.packet = genpkt
	if !sig.Hash.Available() {
		return nil, errors.New("signature uses unknown digest")
	}
	sig.hash = sig.Hash.New()
	return sig, nil
}

func setupDigesters(sigHeader *rpmHeader, payloadDigest io.Writer) ([]*Signature, io.Writer, io.Writer, error) {
	sigs := make([]*Signature, 0)
	headerWriters := make([]io.Writer, 0)
	payloadWriters := []io.Writer{payloadDigest}
	for _, tag := range headerSigTags {
		sig, err := setupDigester(sigHeader, tag)
		if err != nil {
			return nil, nil, nil, err
		} else if sig == nil {
			continue
		}
		sig.HeaderOnly = true
		headerWriters = append(headerWriters, sig.hash)
		sigs = append(sigs, sig)
	}
	for _, tag := range payloadSigTags {
		sig, err := setupDigester(sigHeader, tag)
		if err != nil {
			return nil, nil, nil, err
		} else if sig == nil {
			continue
		}
		sig.HeaderOnly = false
		payloadWriters = append(payloadWriters, sig.hash)
		sigs = append(sigs, sig)
	}
	headerMulti := io.MultiWriter(headerWriters...)
	payloadMulti := io.MultiWriter(payloadWriters...)
	return sigs, headerMulti, payloadMulti, nil
}

func checkSig(sig *Signature, knownKeys openpgp.EntityList) error {
	keys := knownKeys.KeysById(sig.KeyId)
	if keys == nil {
		return fmt.Errorf("keyid %x not found", sig.KeyId)
	}
	key := keys[0]
	sig.Signer = key.Entity
	var err error
	switch pkt := sig.packet.(type) {
	case *packet.Signature:
		err = key.PublicKey.VerifySignature(sig.hash, pkt)
	case *packet.SignatureV3:
		err = key.PublicKey.VerifySignatureV3(sig.hash, pkt)
	}
	if err != nil {
		return err
	}
	sig.packet = nil
	sig.hash = nil
	return nil
}
