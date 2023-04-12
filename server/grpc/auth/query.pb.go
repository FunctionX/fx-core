// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: fx/auth/v1/query.proto

package auth

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type ConvertAddressRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Prefix  string `protobuf:"bytes,2,opt,name=prefix,proto3" json:"prefix,omitempty"`
}

func (m *ConvertAddressRequest) Reset()         { *m = ConvertAddressRequest{} }
func (m *ConvertAddressRequest) String() string { return proto.CompactTextString(m) }
func (*ConvertAddressRequest) ProtoMessage()    {}
func (*ConvertAddressRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bb64561cce2ec376, []int{0}
}
func (m *ConvertAddressRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ConvertAddressRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ConvertAddressRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ConvertAddressRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConvertAddressRequest.Merge(m, src)
}
func (m *ConvertAddressRequest) XXX_Size() int {
	return m.Size()
}
func (m *ConvertAddressRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ConvertAddressRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ConvertAddressRequest proto.InternalMessageInfo

func (m *ConvertAddressRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *ConvertAddressRequest) GetPrefix() string {
	if m != nil {
		return m.Prefix
	}
	return ""
}

type ConvertAddressResponse struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *ConvertAddressResponse) Reset()         { *m = ConvertAddressResponse{} }
func (m *ConvertAddressResponse) String() string { return proto.CompactTextString(m) }
func (*ConvertAddressResponse) ProtoMessage()    {}
func (*ConvertAddressResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bb64561cce2ec376, []int{1}
}
func (m *ConvertAddressResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ConvertAddressResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ConvertAddressResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ConvertAddressResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConvertAddressResponse.Merge(m, src)
}
func (m *ConvertAddressResponse) XXX_Size() int {
	return m.Size()
}
func (m *ConvertAddressResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ConvertAddressResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ConvertAddressResponse proto.InternalMessageInfo

func (m *ConvertAddressResponse) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func init() {
	proto.RegisterType((*ConvertAddressRequest)(nil), "fx.auth.v1.ConvertAddressRequest")
	proto.RegisterType((*ConvertAddressResponse)(nil), "fx.auth.v1.ConvertAddressResponse")
}

func init() { proto.RegisterFile("fx/auth/v1/query.proto", fileDescriptor_bb64561cce2ec376) }

var fileDescriptor_bb64561cce2ec376 = []byte{
	// 314 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x50, 0x3d, 0x4b, 0x03, 0x41,
	0x10, 0xcd, 0x05, 0x8c, 0xb8, 0x85, 0xc5, 0xa1, 0x21, 0x84, 0xb0, 0xe8, 0x61, 0x61, 0x93, 0x1b,
	0x2e, 0xf9, 0x05, 0x6a, 0x21, 0x96, 0xa6, 0xb4, 0xdb, 0xdb, 0xcc, 0x5d, 0x0e, 0xcc, 0xce, 0x65,
	0x77, 0xef, 0x38, 0x91, 0x34, 0x16, 0xd6, 0x82, 0x7f, 0xca, 0x32, 0x60, 0x63, 0x29, 0x89, 0x3f,
	0x44, 0xee, 0x43, 0xfc, 0x40, 0xed, 0x66, 0xde, 0xbc, 0x7d, 0xfb, 0xde, 0x63, 0xdd, 0xa8, 0x00,
	0x91, 0xd9, 0x19, 0xe4, 0x01, 0x2c, 0x32, 0xd4, 0x37, 0x7e, 0xaa, 0xc9, 0x92, 0xcb, 0xa2, 0xc2,
	0x2f, 0x71, 0x3f, 0x0f, 0xfa, 0x5c, 0x92, 0x99, 0x93, 0x81, 0x50, 0x18, 0x84, 0x3c, 0x08, 0xd1,
	0x8a, 0x00, 0x24, 0x25, 0xaa, 0xe6, 0xf6, 0xf7, 0x62, 0x8a, 0xa9, 0x1a, 0xa1, 0x9c, 0x1a, 0x74,
	0x10, 0x13, 0xc5, 0xd7, 0x08, 0x22, 0x4d, 0x40, 0x28, 0x45, 0x56, 0xd8, 0x84, 0x94, 0xa9, 0xaf,
	0xde, 0x05, 0xdb, 0x3f, 0x23, 0x95, 0xa3, 0xb6, 0x27, 0xd3, 0xa9, 0x46, 0x63, 0x26, 0xb8, 0xc8,
	0xd0, 0x58, 0xb7, 0xc7, 0xb6, 0x45, 0x8d, 0xf4, 0x9c, 0x03, 0xe7, 0x78, 0x67, 0xf2, 0xb1, 0xba,
	0x5d, 0xd6, 0x49, 0x35, 0x46, 0x49, 0xd1, 0x6b, 0x57, 0x87, 0x66, 0xf3, 0x46, 0xac, 0xfb, 0x53,
	0xca, 0xa4, 0xa4, 0x0c, 0xfe, 0xad, 0x35, 0xba, 0x77, 0xd8, 0xd6, 0x65, 0x19, 0xd7, 0x5d, 0xb2,
	0xdd, 0xef, 0xaf, 0xdd, 0x43, 0xff, 0x33, 0xbb, 0xff, 0xab, 0xc9, 0xbe, 0xf7, 0x1f, 0xa5, 0xfe,
	0xdc, 0x3b, 0xba, 0x7b, 0x7e, 0x7b, 0x6c, 0x73, 0x77, 0x00, 0x5f, 0x2a, 0x0e, 0x51, 0xce, 0xc6,
	0x23, 0xb8, 0x6d, 0x7c, 0x2c, 0x4f, 0xcf, 0x9f, 0xd6, 0xdc, 0x59, 0xad, 0xb9, 0xf3, 0xba, 0xe6,
	0xce, 0xc3, 0x86, 0xb7, 0x56, 0x1b, 0xde, 0x7a, 0xd9, 0xf0, 0xd6, 0xd5, 0x30, 0x4e, 0xec, 0x2c,
	0x0b, 0x7d, 0x49, 0x73, 0x88, 0x32, 0x25, 0xcb, 0xf6, 0x0a, 0x88, 0x8a, 0xa1, 0x24, 0x8d, 0x60,
	0x50, 0xe7, 0xa8, 0x21, 0xd6, 0xa9, 0xac, 0xc4, 0xc3, 0x4e, 0xd5, 0xeb, 0xf8, 0x3d, 0x00, 0x00,
	0xff, 0xff, 0x15, 0x84, 0xc8, 0xd9, 0xd1, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	ConvertAddress(ctx context.Context, in *ConvertAddressRequest, opts ...grpc.CallOption) (*ConvertAddressResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) ConvertAddress(ctx context.Context, in *ConvertAddressRequest, opts ...grpc.CallOption) (*ConvertAddressResponse, error) {
	out := new(ConvertAddressResponse)
	err := c.cc.Invoke(ctx, "/fx.auth.v1.Query/ConvertAddress", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	ConvertAddress(context.Context, *ConvertAddressRequest) (*ConvertAddressResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) ConvertAddress(ctx context.Context, req *ConvertAddressRequest) (*ConvertAddressResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConvertAddress not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_ConvertAddress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConvertAddressRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ConvertAddress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fx.auth.v1.Query/ConvertAddress",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ConvertAddress(ctx, req.(*ConvertAddressRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "fx.auth.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ConvertAddress",
			Handler:    _Query_ConvertAddress_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "fx/auth/v1/query.proto",
}

func (m *ConvertAddressRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ConvertAddressRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ConvertAddressRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Prefix) > 0 {
		i -= len(m.Prefix)
		copy(dAtA[i:], m.Prefix)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Prefix)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *ConvertAddressResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ConvertAddressResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ConvertAddressResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ConvertAddressRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	l = len(m.Prefix)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *ConvertAddressResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ConvertAddressRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ConvertAddressRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ConvertAddressRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Prefix", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Prefix = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ConvertAddressResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ConvertAddressResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ConvertAddressResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery = fmt.Errorf("proto: unexpected end of group")
)