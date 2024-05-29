// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: fx/legacy/other/query.proto

package v1

import (
	context "context"
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
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

// Deprecated: GasPriceRequest
type GasPriceRequest struct {
}

func (m *GasPriceRequest) Reset()         { *m = GasPriceRequest{} }
func (m *GasPriceRequest) String() string { return proto.CompactTextString(m) }
func (*GasPriceRequest) ProtoMessage()    {}
func (*GasPriceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_853fc8108b6a0fa5, []int{0}
}
func (m *GasPriceRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GasPriceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GasPriceRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GasPriceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GasPriceRequest.Merge(m, src)
}
func (m *GasPriceRequest) XXX_Size() int {
	return m.Size()
}
func (m *GasPriceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GasPriceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GasPriceRequest proto.InternalMessageInfo

// Deprecated: GasPriceResponse
type GasPriceResponse struct {
	GasPrices github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,1,rep,name=gas_prices,json=gasPrices,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"gas_prices" yaml:"gas_prices"`
}

func (m *GasPriceResponse) Reset()         { *m = GasPriceResponse{} }
func (m *GasPriceResponse) String() string { return proto.CompactTextString(m) }
func (*GasPriceResponse) ProtoMessage()    {}
func (*GasPriceResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_853fc8108b6a0fa5, []int{1}
}
func (m *GasPriceResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GasPriceResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GasPriceResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GasPriceResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GasPriceResponse.Merge(m, src)
}
func (m *GasPriceResponse) XXX_Size() int {
	return m.Size()
}
func (m *GasPriceResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GasPriceResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GasPriceResponse proto.InternalMessageInfo

func (m *GasPriceResponse) GetGasPrices() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.GasPrices
	}
	return nil
}

func init() {
	proto.RegisterType((*GasPriceRequest)(nil), "fx.other.GasPriceRequest")
	proto.RegisterType((*GasPriceResponse)(nil), "fx.other.GasPriceResponse")
}

func init() { proto.RegisterFile("fx/legacy/other/query.proto", fileDescriptor_853fc8108b6a0fa5) }

var fileDescriptor_853fc8108b6a0fa5 = []byte{
	// 381 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x92, 0x3f, 0x4e, 0xe3, 0x40,
	0x14, 0xc6, 0xed, 0x5d, 0xed, 0x2a, 0x3b, 0x5b, 0xec, 0xc6, 0xbb, 0x2b, 0x6d, 0x9c, 0x95, 0xb3,
	0x72, 0x95, 0x26, 0x33, 0x72, 0xe8, 0xa0, 0x0b, 0x02, 0x5a, 0x88, 0x44, 0x43, 0x01, 0x8c, 0x87,
	0xf1, 0xc4, 0x22, 0xf1, 0x73, 0x3c, 0x63, 0xcb, 0xae, 0x90, 0x38, 0x01, 0x52, 0x6e, 0xc1, 0x49,
	0x42, 0x17, 0x89, 0x86, 0x2a, 0xa0, 0x84, 0x13, 0x70, 0x02, 0xe4, 0x3f, 0x21, 0x08, 0x51, 0x52,
	0xcd, 0xd3, 0xf7, 0x66, 0x3e, 0xfd, 0xde, 0xf7, 0x06, 0x35, 0xbd, 0x94, 0x0c, 0xb9, 0xa0, 0x2c,
	0x23, 0xa0, 0x06, 0x3c, 0x22, 0xe3, 0x98, 0x47, 0x19, 0x0e, 0x23, 0x50, 0x60, 0xd4, 0xbc, 0x14,
	0x17, 0xaa, 0x69, 0x31, 0x90, 0x23, 0x90, 0xc4, 0xa5, 0x92, 0x93, 0xc4, 0x71, 0xb9, 0xa2, 0x0e,
	0x61, 0xe0, 0x07, 0xe5, 0x4d, 0xf3, 0xb7, 0x00, 0x01, 0x45, 0x49, 0xf2, 0xaa, 0x52, 0xff, 0x09,
	0x00, 0x31, 0xe4, 0x84, 0x86, 0x3e, 0xa1, 0x41, 0x00, 0x8a, 0x2a, 0x1f, 0x02, 0x59, 0x76, 0xed,
	0x3a, 0xfa, 0xb1, 0x47, 0xe5, 0x7e, 0xe4, 0x33, 0xde, 0xe7, 0xe3, 0x98, 0x4b, 0x65, 0x4f, 0x74,
	0xf4, 0x73, 0xad, 0xc9, 0x10, 0x02, 0xc9, 0x8d, 0x0b, 0x84, 0x04, 0x95, 0x27, 0x61, 0x2e, 0xca,
	0xbf, 0xfa, 0xff, 0xcf, 0xed, 0xef, 0xdd, 0x06, 0x2e, 0x81, 0x70, 0x0e, 0x84, 0x2b, 0x20, 0xbc,
	0x0d, 0x7e, 0xd0, 0xdb, 0x99, 0xce, 0x5b, 0xda, 0xd3, 0xbc, 0x55, 0xcf, 0xe8, 0x68, 0xb8, 0x69,
	0xaf, 0x9f, 0xda, 0xd7, 0xf7, 0xad, 0xb6, 0xf0, 0xd5, 0x20, 0x76, 0x31, 0x83, 0x11, 0xa9, 0x46,
	0x2a, 0x8f, 0x8e, 0x3c, 0x3b, 0x27, 0x2a, 0x0b, 0xb9, 0x2c, 0x5c, 0x64, 0xff, 0x9b, 0xa8, 0x38,
	0x64, 0xf7, 0x46, 0x47, 0x5f, 0x0e, 0xf2, 0x58, 0x8c, 0x53, 0x84, 0x76, 0xd3, 0x15, 0xa0, 0xd1,
	0xc0, 0xab, 0x7c, 0xf0, 0x9b, 0x41, 0x4c, 0xf3, 0xbd, 0x56, 0x39, 0x8f, 0xdd, 0xbc, 0xbc, 0x7d,
	0x9c, 0x7c, 0xfa, 0x63, 0xfc, 0x22, 0x5e, 0x5a, 0x85, 0xfe, 0x02, 0x69, 0x1c, 0xa3, 0xda, 0xc7,
	0xf9, 0x97, 0xe6, 0x89, 0xb3, 0xf6, 0xef, 0x1d, 0x4e, 0x17, 0x96, 0x3e, 0x5b, 0x58, 0xfa, 0xc3,
	0xc2, 0xd2, 0xaf, 0x96, 0x96, 0x36, 0x5b, 0x5a, 0xda, 0xdd, 0xd2, 0xd2, 0x8e, 0xb6, 0x5e, 0x45,
	0xe3, 0xc5, 0x01, 0xcb, 0x57, 0x95, 0x12, 0x2f, 0xed, 0x30, 0x88, 0x38, 0x91, 0x3c, 0x4a, 0x72,
	0xd0, 0x28, 0x64, 0xb9, 0x5b, 0x61, 0xb6, 0xfa, 0x38, 0x89, 0xe3, 0x7e, 0x2d, 0x56, 0xba, 0xf1,
	0x1c, 0x00, 0x00, 0xff, 0xff, 0xd6, 0xd0, 0x33, 0x44, 0x4f, 0x02, 0x00, 0x00,
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
	// Deprecated: please use cosmos.base.node.v1beta1.Service.Config
	FxGasPrice(ctx context.Context, in *GasPriceRequest, opts ...grpc.CallOption) (*GasPriceResponse, error)
	// Deprecated: please use cosmos.base.node.v1beta1.Service.Config
	GasPrice(ctx context.Context, in *GasPriceRequest, opts ...grpc.CallOption) (*GasPriceResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) FxGasPrice(ctx context.Context, in *GasPriceRequest, opts ...grpc.CallOption) (*GasPriceResponse, error) {
	out := new(GasPriceResponse)
	err := c.cc.Invoke(ctx, "/fx.other.Query/FxGasPrice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) GasPrice(ctx context.Context, in *GasPriceRequest, opts ...grpc.CallOption) (*GasPriceResponse, error) {
	out := new(GasPriceResponse)
	err := c.cc.Invoke(ctx, "/fx.other.Query/GasPrice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// Deprecated: please use cosmos.base.node.v1beta1.Service.Config
	FxGasPrice(context.Context, *GasPriceRequest) (*GasPriceResponse, error)
	// Deprecated: please use cosmos.base.node.v1beta1.Service.Config
	GasPrice(context.Context, *GasPriceRequest) (*GasPriceResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) FxGasPrice(ctx context.Context, req *GasPriceRequest) (*GasPriceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FxGasPrice not implemented")
}
func (*UnimplementedQueryServer) GasPrice(ctx context.Context, req *GasPriceRequest) (*GasPriceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GasPrice not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_FxGasPrice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GasPriceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).FxGasPrice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fx.other.Query/FxGasPrice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).FxGasPrice(ctx, req.(*GasPriceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_GasPrice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GasPriceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).GasPrice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fx.other.Query/GasPrice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).GasPrice(ctx, req.(*GasPriceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "fx.other.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FxGasPrice",
			Handler:    _Query_FxGasPrice_Handler,
		},
		{
			MethodName: "GasPrice",
			Handler:    _Query_GasPrice_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "fx/legacy/other/query.proto",
}

func (m *GasPriceRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GasPriceRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GasPriceRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *GasPriceResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GasPriceResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GasPriceResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.GasPrices) > 0 {
		for iNdEx := len(m.GasPrices) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.GasPrices[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
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
func (m *GasPriceRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *GasPriceResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.GasPrices) > 0 {
		for _, e := range m.GasPrices {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GasPriceRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: GasPriceRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GasPriceRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
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
func (m *GasPriceResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: GasPriceResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GasPriceResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GasPrices", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.GasPrices = append(m.GasPrices, types.Coin{})
			if err := m.GasPrices[len(m.GasPrices)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
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