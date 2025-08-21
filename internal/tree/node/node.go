package node

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Type uint8

const (
	TypeInternal Type = iota
	TypeLeaf
)

type Node interface {
	Type() Type
	Search(key []byte) (int, bool)
	Key(i int) ([]byte, error)
	Split(maxSize int) []Node
}

type Header struct {
	typ   Type
	meta  uint8
	count uint16
}

func Serialize(node Node) ([]byte, error) {
	buf := new(bytes.Buffer)
	switch n := node.(type) {
	case *Internal:
		if err := serializeInternalNode(buf, n); err != nil {
			return nil, fmt.Errorf("node: failed to serialize internal node: %w", err)
		}
	case *Leaf:
		if err := serializeLeafNode(buf, n); err != nil {
			return nil, fmt.Errorf("node: failed to serialize leaf node: %w", err)
		}
	default:
		return nil, fmt.Errorf("node: failed to serialize node: invalid node type '%T'", n)
	}

	return buf.Bytes(), nil
}

func Deserialize(d []byte) (Node, error) {
	r := bytes.NewReader(d)

	head, err := deserializeHeader(r)
	if err != nil {
		return nil, err
	}

	switch head.typ {
	case TypeInternal:
		n, err := deserializeInternalNode(r, head)
		if err != nil {
			return nil, fmt.Errorf("node: failed to deserialize internal node: %w", err)
		}
		return n, nil
	case TypeLeaf:
		n, err := deserializeLeafNode(r, head)
		if err != nil {
			return nil, fmt.Errorf("node: failed to deserialize leaf node: %w", err)
		}
		return n, nil
	default:
		return nil, fmt.Errorf("node: failed to deserialize: invalid node type '%x'", head.typ)
	}
}

func serializeHeader(w io.Writer, head Header) error {
	if err := binary.Write(w, binary.LittleEndian, head.typ); err != nil {
		return fmt.Errorf("failed to write node type: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, head.count); err != nil {
		return fmt.Errorf("failed to write entry count: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, head.meta); err != nil {
		return fmt.Errorf("failed to write meta data: %w", err)
	}
	return nil
}

func deserializeHeader(r io.Reader) (Header, error) {
	var head Header
	if err := binary.Read(r, binary.LittleEndian, &head.typ); err != nil {
		return Header{}, fmt.Errorf("failed to read node type: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &head.count); err != nil {
		return Header{}, fmt.Errorf("failed to read entry count: %w", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &head.meta); err != nil {
		return Header{}, fmt.Errorf("failed to read meta data: %w", err)
	}
	return head, nil
}

func serializeLeafNode(w io.Writer, n *Leaf) error {
	if err := serializeHeader(w, n.head); err != nil {
		return fmt.Errorf("node: failed to serialize header: %w", err)
	}

	for i := range n.head.count {
		k, v := n.keys[i], n.vals[i]
		if err := binary.Write(w, binary.LittleEndian, uint16(len(k))); err != nil {
			return fmt.Errorf("failed to write key length: %w", err)
		}
		if err := binary.Write(w, binary.LittleEndian, uint16(len(v))); err != nil {
			return fmt.Errorf("failed to write value length: %w", err)
		}
		if err := binary.Write(w, binary.LittleEndian, k); err != nil {
			return fmt.Errorf("failed to write key: %w", err)
		}
		if err := binary.Write(w, binary.LittleEndian, v); err != nil {
			return fmt.Errorf("failed to write value: %w", err)
		}
	}
	return nil
}

func deserializeLeafNode(r io.Reader, head Header) (*Leaf, error) {
	node := &Leaf{head: head}
	node.keys = make([][]byte, head.count)
	node.vals = make([][]byte, head.count)

	var kLen, vLen uint16
	for i := range head.count {
		if err := binary.Read(r, binary.LittleEndian, &kLen); err != nil {
			return nil, fmt.Errorf("node: failed to read key length: %w", err)
		}
		if err := binary.Read(r, binary.LittleEndian, &vLen); err != nil {
			return nil, fmt.Errorf("node: failed to read value length: %w", err)
		}

		node.keys[i] = make([]byte, kLen)
		if err := binary.Read(r, binary.LittleEndian, &node.keys[i]); err != nil {
			return nil, fmt.Errorf("node: failed to read key: %w", err)
		}
		node.vals[i] = make([]byte, vLen)
		if err := binary.Read(r, binary.LittleEndian, &node.vals[i]); err != nil {
			return nil, fmt.Errorf("node: failed to read value: %w", err)
		}
	}
	return node, nil
}

func serializeInternalNode(w io.Writer, n *Internal) error {
	if err := serializeHeader(w, n.head); err != nil {
		return fmt.Errorf("node: failed to serialize header: %w", err)
	}

	for i := range n.head.count {
		k, p := n.keys[i], n.ptrs[i]
		if err := binary.Write(w, binary.LittleEndian, uint16(len(k))); err != nil {
			return fmt.Errorf("failed to write key length: %w", err)
		}
		if err := binary.Write(w, binary.LittleEndian, k); err != nil {
			return fmt.Errorf("failed to write key: %w", err)
		}
		if err := binary.Write(w, binary.LittleEndian, p); err != nil {
			return fmt.Errorf("failed to write pointer: %w", err)
		}
	}
	return nil
}

func deserializeInternalNode(r io.Reader, head Header) (*Internal, error) {
	node := &Internal{head: head}
	node.keys = make([][]byte, head.count)
	node.ptrs = make([]uint64, head.count)

	for i := range head.count {
		var kLen uint16
		if err := binary.Read(r, binary.LittleEndian, &kLen); err != nil {
			return nil, fmt.Errorf("node: failed to read key length: %w", err)
		}

		node.keys[i] = make([]byte, kLen)
		if err := binary.Read(r, binary.LittleEndian, &node.keys[i]); err != nil {
			return nil, fmt.Errorf("node: failed to read key: %w", err)
		}
		if err := binary.Read(r, binary.LittleEndian, &node.ptrs[i]); err != nil {
			return nil, fmt.Errorf("node: failed to read pointer: %w", err)
		}
	}
	return node, nil
}
