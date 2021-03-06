// Code generated by protoc-gen-go. DO NOT EDIT.
// source: db_create.proto

// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package yaml

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	admin "github.com/m3db/m3/src/query/generated/proto/admin"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type DatabaseCreateRequestYaml struct {
	Operation            string                       `protobuf:"bytes,1,opt,name=operation,proto3" json:"operation,omitempty"`
	Request              *admin.DatabaseCreateRequest `protobuf:"bytes,2,opt,name=request,proto3" json:"request,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                     `json:"-"`
	XXX_unrecognized     []byte                       `json:"-"`
	XXX_sizecache        int32                        `json:"-"`
}

func (m *DatabaseCreateRequestYaml) Reset()         { *m = DatabaseCreateRequestYaml{} }
func (m *DatabaseCreateRequestYaml) String() string { return proto.CompactTextString(m) }
func (*DatabaseCreateRequestYaml) ProtoMessage()    {}
func (*DatabaseCreateRequestYaml) Descriptor() ([]byte, []int) {
	return fileDescriptor_57e276f15713f139, []int{0}
}

func (m *DatabaseCreateRequestYaml) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DatabaseCreateRequestYaml.Unmarshal(m, b)
}
func (m *DatabaseCreateRequestYaml) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DatabaseCreateRequestYaml.Marshal(b, m, deterministic)
}
func (m *DatabaseCreateRequestYaml) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DatabaseCreateRequestYaml.Merge(m, src)
}
func (m *DatabaseCreateRequestYaml) XXX_Size() int {
	return xxx_messageInfo_DatabaseCreateRequestYaml.Size(m)
}
func (m *DatabaseCreateRequestYaml) XXX_DiscardUnknown() {
	xxx_messageInfo_DatabaseCreateRequestYaml.DiscardUnknown(m)
}

var xxx_messageInfo_DatabaseCreateRequestYaml proto.InternalMessageInfo

func (m *DatabaseCreateRequestYaml) GetOperation() string {
	if m != nil {
		return m.Operation
	}
	return ""
}

func (m *DatabaseCreateRequestYaml) GetRequest() *admin.DatabaseCreateRequest {
	if m != nil {
		return m.Request
	}
	return nil
}

func init() {
	proto.RegisterType((*DatabaseCreateRequestYaml)(nil), "yaml.DatabaseCreateRequestYaml")
}

func init() { proto.RegisterFile("db_create.proto", fileDescriptor_57e276f15713f139) }

var fileDescriptor_57e276f15713f139 = []byte{
	// 177 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x8d, 0xb1, 0xca, 0xc2, 0x30,
	0x14, 0x46, 0xe9, 0xcf, 0x8f, 0xd2, 0x38, 0x08, 0x9d, 0xaa, 0x74, 0x28, 0x4e, 0x9d, 0x72, 0xc1,
	0x82, 0xbb, 0xe8, 0x13, 0x74, 0x73, 0x92, 0x9b, 0xe6, 0x52, 0x0b, 0x4d, 0xd3, 0xa6, 0x37, 0x43,
	0xdf, 0x5e, 0x8c, 0x15, 0x17, 0xd7, 0x8f, 0xf3, 0x9d, 0x23, 0xb6, 0x5a, 0xdd, 0x6b, 0x47, 0xc8,
	0x24, 0x07, 0x67, 0xd9, 0x26, 0xff, 0x33, 0x9a, 0x6e, 0x7f, 0x6e, 0x5a, 0x7e, 0x78, 0x25, 0x6b,
	0x6b, 0xc0, 0x94, 0x5a, 0x81, 0x29, 0x61, 0x72, 0x35, 0x8c, 0x9e, 0xdc, 0x0c, 0x0d, 0xf5, 0xe4,
	0x90, 0x49, 0x43, 0xf8, 0x00, 0x6a, 0xd3, 0xf6, 0xa0, 0x91, 0x51, 0xe1, 0xb4, 0x88, 0x0e, 0xa3,
	0xd8, 0x5d, 0x97, 0xe5, 0x12, 0x02, 0x15, 0x8d, 0x9e, 0x26, 0xbe, 0xa1, 0xe9, 0x92, 0x4c, 0xc4,
	0x76, 0x78, 0x39, 0x5a, 0xdb, 0xa7, 0x51, 0x1e, 0x15, 0x71, 0xf5, 0x1d, 0x92, 0x93, 0x58, 0xbb,
	0x37, 0x9c, 0xfe, 0xe5, 0x51, 0xb1, 0x39, 0x66, 0x32, 0x24, 0xe4, 0x4f, 0x61, 0xf5, 0x81, 0xd5,
	0x2a, 0x94, 0xcb, 0x67, 0x00, 0x00, 0x00, 0xff, 0xff, 0xbe, 0x30, 0x42, 0x7b, 0xd5, 0x00, 0x00,
	0x00,
}
