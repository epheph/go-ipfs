// Package iface defines IPFS Core API which is a set of interfaces used to
// interact with IPFS nodes.
package iface

import (
	"context"
	"errors"
	"io"

	options "github.com/ipfs/go-ipfs/core/coreapi/interface/options"

	ipld "gx/ipfs/QmNwUEK7QbwSqyKBu3mMtToo8SUc6wQJ7gdZq4gGGJqfnf/go-ipld-format"
	cid "gx/ipfs/QmeSrf6pzut73u6zLQkRFQ3ygt3k6XFT2kjdYP8Tnkwwyg/go-cid"
)

// Path is a generic wrapper for paths used in the API. A path can be resolved
// to a CID using one of Resolve functions in the API.
type Path interface {
	String() string
	Cid() *cid.Cid
	Root() *cid.Cid
	Resolved() bool
}

// TODO: should we really copy these?
//       if we didn't, godoc would generate nice links straight to go-ipld-format
type Node ipld.Node
type Link ipld.Link

type Reader interface {
	io.ReadSeeker
	io.Closer
}

// CoreAPI defines an unified interface to IPFS for Go programs.
type CoreAPI interface {
	// Unixfs returns an implementation of Unixfs API
	Unixfs() UnixfsAPI
	Dag() DagAPI

	// ResolvePath resolves the path using Unixfs resolver
	ResolvePath(context.Context, Path) (Path, error)

	// ResolveNode resolves the path (if not resolved already) using Unixfs
	// resolver, gets and returns the resolved Node
	ResolveNode(context.Context, Path) (Node, error)
}

// UnixfsAPI is the basic interface to immutable files in IPFS
type UnixfsAPI interface {
	// Add imports the data from the reader into merkledag file
	Add(context.Context, io.Reader) (Path, error)

	// Cat returns a reader for the file
	Cat(context.Context, Path) (Reader, error)

	// Ls returns the list of links in a directory
	Ls(context.Context, Path) ([]*Link, error)
}

// DagAPI specifies the interface to IPLD
type DagAPI interface {
	// Put inserts data using specified format and input encoding.
	// Unless used with WithCodec or WithHash, the defaults "dag-cbor" and
	// "sha256" are used.
	Put(ctx context.Context, src io.Reader, opts ...options.DagPutOption) (Path, error)

	// WithInputEnc is an option for Put which specifies the input encoding of the
	// data. Default is "json", most formats/codecs support "raw"
	WithInputEnc(enc string) options.DagPutOption

	// WithCodec is an option for Put which specifies the multicodec to use to
	// serialize the object. Default is cid.DagCBOR (0x71)
	WithCodec(codec uint64) options.DagPutOption

	// WithHash is an option for Put which specifies the multihash settings to use
	// when hashing the object. Default is based on the codec used
	// (mh.SHA2_256 (0x12) for DagCBOR). If mhLen is set to -1, default length for
	// the hash will be used
	WithHash(mhType uint64, mhLen int) options.DagPutOption

	// Get attempts to resolve and get the node specified by the path
	Get(ctx context.Context, path Path) (Node, error)

	// Tree returns list of paths within a node specified by the path.
	Tree(ctx context.Context, path Path, opts ...options.DagTreeOption) ([]Path, error)

	// WithDepth is an option for Tree which specifies maximum depth of the
	// returned tree. Default is -1 (no depth limit)
	WithDepth(depth int) options.DagTreeOption
}

// ObjectAPI specifies the interface to MerkleDAG and contains useful utilities
// for manipulating MerkleDAG data structures.
type ObjectAPI interface {
	// New creates new, empty (by default) dag-node.
	New(context.Context, ...options.ObjectNewOption) (Node, error)

	// WithType is an option for New which allows to change the type of created
	// dag node.
	//
	// Supported types:
	// * 'empty' - Empty node
	// * 'unixfs-dir' - Empty UnixFS directory
	WithType(string) options.ObjectNewOption

	// Put imports the node into merkledag
	Put(context.Context, Node) (Path, error)

	// Get returns the node for the path
	Get(context.Context, Path) (Node, error)

	// Data returns reader for data of the node
	Data(context.Context, Path) (io.Reader, error)

	// Links returns lint or links the node contains
	Links(context.Context, Path) ([]*Link, error)

	// Stat returns information about the node
	Stat(context.Context, Path) (*ObjectStat, error)

	// AddLink adds a link under the specified path. child path can point to a
	// subdirectory within the patent which must be present (can be overridden
	// with WithCreate option).
	AddLink(ctx context.Context, base Path, name string, child Path, opts ...options.ObjectAddLinkOption) (Node, error)

	// WithCreate is an option for AddLink which specifies whether create required
	// directories for the child
	WithCreate(create bool) options.ObjectAddLinkOption

	// RmLink removes a link from the node
	RmLink(ctx context.Context, base Path, link string) (Node, error)

	// AppendData appends data to the node
	AppendData(context.Context, Path, io.Reader) (Node, error)

	// SetData sets the data contained in the node
	SetData(context.Context, Path, io.Reader) (Node, error)
}

// ObjectStat provides information about dag nodes
type ObjectStat struct {
	// Cid is the CID of the node
	Cid *cid.Cid

	// NumLinks is number of links the node contains
	NumLinks int

	// BlockSize is size of the raw serialized node
	BlockSize int

	// LinksSize is size of the links block section
	LinksSize int

	// DataSize is the size of data block section
	DataSize int

	// CumulativeSize is size of node
	CumulativeSize int
}

var ErrIsDir = errors.New("object is a directory")
var ErrOffline = errors.New("can't resolve, ipfs node is offline")
