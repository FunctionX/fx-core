// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: fx/ibc/applications/transfer/v1/ibc_legacy.proto

package legacy

import (
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	_ "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	types1 "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
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

// MsgTransfer defines a msg to transfer fungible tokens (i.e Coins) between
// ICS20 enabled chains. See ICS Spec here:
// https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#data-structures
type MsgTransfer struct {
	// the port on which the packet will be sent
	SourcePort string `protobuf:"bytes,1,opt,name=source_port,json=sourcePort,proto3" json:"source_port,omitempty" yaml:"source_port"`
	// the channel by which the packet will be sent
	SourceChannel string `protobuf:"bytes,2,opt,name=source_channel,json=sourceChannel,proto3" json:"source_channel,omitempty" yaml:"source_channel"`
	// the tokens to be transferred
	Token types.Coin `protobuf:"bytes,3,opt,name=token,proto3" json:"token"`
	// the sender address
	Sender string `protobuf:"bytes,4,opt,name=sender,proto3" json:"sender,omitempty"`
	// the recipient address on the destination chain
	Receiver string `protobuf:"bytes,5,opt,name=receiver,proto3" json:"receiver,omitempty"`
	// Timeout height relative to the current block height.
	// The timeout is disabled when set to 0.
	TimeoutHeight types1.Height `protobuf:"bytes,6,opt,name=timeout_height,json=timeoutHeight,proto3" json:"timeout_height" yaml:"timeout_height"`
	// Timeout timestamp (in nanoseconds) relative to the current block timestamp.
	// The timeout is disabled when set to 0.
	TimeoutTimestamp uint64 `protobuf:"varint,7,opt,name=timeout_timestamp,json=timeoutTimestamp,proto3" json:"timeout_timestamp,omitempty" yaml:"timeout_timestamp"`
	// the router is hook destination chain
	Router string `protobuf:"bytes,8,opt,name=router,proto3" json:"router,omitempty"`
	// the tokens to be destination fee
	Fee types.Coin `protobuf:"bytes,9,opt,name=fee,proto3" json:"fee"`
	// optional memo
	Memo string `protobuf:"bytes,10,opt,name=memo,proto3" json:"memo,omitempty"`
}

func (m *MsgTransfer) Reset()         { *m = MsgTransfer{} }
func (m *MsgTransfer) String() string { return proto.CompactTextString(m) }
func (*MsgTransfer) ProtoMessage()    {}
func (*MsgTransfer) Descriptor() ([]byte, []int) {
	return fileDescriptor_94c7270e688125fc, []int{0}
}
func (m *MsgTransfer) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgTransfer) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgTransfer.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgTransfer) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgTransfer.Merge(m, src)
}
func (m *MsgTransfer) XXX_Size() int {
	return m.Size()
}
func (m *MsgTransfer) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgTransfer.DiscardUnknown(m)
}

var xxx_messageInfo_MsgTransfer proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgTransfer)(nil), "fx.ibc.applications.transfer.v1.MsgTransfer")
}

func init() {
	proto.RegisterFile("fx/ibc/applications/transfer/v1/ibc_legacy.proto", fileDescriptor_94c7270e688125fc)
}

var fileDescriptor_94c7270e688125fc = []byte{
	// 498 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x52, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0xb5, 0x69, 0x1a, 0xd2, 0x8d, 0x5a, 0xc1, 0x0a, 0xaa, 0x6d, 0x04, 0x76, 0x64, 0x09, 0x29,
	0x1c, 0xd8, 0xc5, 0x20, 0x84, 0xd4, 0x13, 0x4a, 0x2e, 0x70, 0x40, 0x42, 0x56, 0x4f, 0x5c, 0x82,
	0xbd, 0x8c, 0x9d, 0x15, 0xb1, 0xd7, 0x5a, 0x6f, 0xa2, 0xe4, 0x0f, 0xe0, 0xc6, 0x27, 0xf4, 0x73,
	0x7a, 0xec, 0x91, 0x53, 0x84, 0x92, 0x0b, 0xe7, 0x7c, 0x01, 0x5a, 0xaf, 0x53, 0x9a, 0x0b, 0xea,
	0xc9, 0x33, 0xf3, 0xde, 0x9b, 0xf1, 0xbc, 0x1d, 0xf4, 0x32, 0x5d, 0x30, 0x91, 0x70, 0x16, 0x97,
	0xe5, 0x54, 0xf0, 0x58, 0x0b, 0x59, 0x54, 0x4c, 0xab, 0xb8, 0xa8, 0x52, 0x50, 0x6c, 0x1e, 0x1a,
	0x70, 0x3c, 0x85, 0x2c, 0xe6, 0x4b, 0x5a, 0x2a, 0xa9, 0x25, 0xf6, 0xd3, 0x05, 0x15, 0x09, 0xa7,
	0xb7, 0x15, 0x74, 0xa7, 0xa0, 0xf3, 0xb0, 0xe7, 0x71, 0x59, 0xe5, 0xb2, 0x62, 0x49, 0x5c, 0x01,
	0x9b, 0x87, 0x09, 0xe8, 0x38, 0x64, 0x5c, 0x8a, 0xc2, 0x36, 0xe8, 0x3d, 0xca, 0x64, 0x26, 0xeb,
	0x90, 0x99, 0xa8, 0xa9, 0x3e, 0xfb, 0xef, 0x5f, 0xe8, 0x45, 0x43, 0xf3, 0x0d, 0x8d, 0x4b, 0x05,
	0x8c, 0x4f, 0x05, 0x14, 0xda, 0x80, 0x36, 0xb2, 0x84, 0xe0, 0x47, 0x0b, 0x75, 0x3f, 0x56, 0xd9,
	0x45, 0x23, 0xc6, 0x6f, 0x51, 0xb7, 0x92, 0x33, 0xc5, 0x61, 0x5c, 0x4a, 0xa5, 0x89, 0xdb, 0x77,
	0x07, 0x47, 0xc3, 0xd3, 0xed, 0xca, 0xc7, 0xcb, 0x38, 0x9f, 0x9e, 0x07, 0xb7, 0xc0, 0x20, 0x42,
	0x36, 0xfb, 0x24, 0x95, 0xc6, 0xef, 0xd0, 0x49, 0x83, 0xf1, 0x49, 0x5c, 0x14, 0x30, 0x25, 0xf7,
	0x6a, 0xed, 0xd9, 0x76, 0xe5, 0x3f, 0xde, 0xd3, 0x36, 0x78, 0x10, 0x1d, 0xdb, 0xc2, 0xc8, 0xe6,
	0xf8, 0x0d, 0x3a, 0xd4, 0xf2, 0x1b, 0x14, 0xe4, 0xa0, 0xef, 0x0e, 0xba, 0xaf, 0xce, 0xa8, 0x35,
	0x86, 0x1a, 0x63, 0x68, 0x63, 0x0c, 0x1d, 0x49, 0x51, 0x0c, 0x5b, 0x57, 0x2b, 0xdf, 0x89, 0x2c,
	0x1b, 0x9f, 0xa2, 0x76, 0x05, 0xc5, 0x57, 0x50, 0xa4, 0x65, 0x06, 0x46, 0x4d, 0x86, 0x7b, 0xa8,
	0xa3, 0x80, 0x83, 0x98, 0x83, 0x22, 0x87, 0x35, 0x72, 0x93, 0xe3, 0x2f, 0xe8, 0x44, 0x8b, 0x1c,
	0xe4, 0x4c, 0x8f, 0x27, 0x20, 0xb2, 0x89, 0x26, 0xed, 0x7a, 0x66, 0xaf, 0x7e, 0x2a, 0xe3, 0x17,
	0x6d, 0x5c, 0x9a, 0x87, 0xf4, 0x7d, 0xcd, 0x18, 0x3e, 0x35, 0x43, 0xff, 0x2d, 0xb3, 0xaf, 0x0f,
	0xa2, 0xe3, 0xa6, 0x60, 0xd9, 0xf8, 0x03, 0x7a, 0xb8, 0x63, 0x98, 0x6f, 0xa5, 0xe3, 0xbc, 0x24,
	0xf7, 0xfb, 0xee, 0xa0, 0x35, 0x7c, 0xb2, 0x5d, 0xf9, 0x64, 0xbf, 0xc9, 0x0d, 0x25, 0x88, 0x1e,
	0x34, 0xb5, 0x8b, 0x5d, 0xc9, 0x2c, 0xa8, 0xe4, 0x4c, 0x83, 0x22, 0x1d, 0xbb, 0xa0, 0xcd, 0x70,
	0x88, 0x0e, 0x52, 0x00, 0x72, 0x74, 0x37, 0xb7, 0x0c, 0x17, 0x63, 0xd4, 0xca, 0x21, 0x97, 0x04,
	0xd5, 0x8d, 0xea, 0xf8, 0xbc, 0xf3, 0xfd, 0xd2, 0x77, 0xfe, 0x5c, 0xfa, 0xce, 0x70, 0x74, 0xb5,
	0xf6, 0xdc, 0xeb, 0xb5, 0xe7, 0xfe, 0x5e, 0x7b, 0xee, 0xcf, 0x8d, 0xe7, 0x5c, 0x6f, 0x3c, 0xe7,
	0xd7, 0xc6, 0x73, 0x3e, 0x3f, 0xcf, 0x84, 0x9e, 0xcc, 0x12, 0xca, 0x65, 0xce, 0xd2, 0x59, 0xc1,
	0xcd, 0xc5, 0x2d, 0x58, 0xba, 0x78, 0x51, 0x9f, 0x96, 0x5e, 0x96, 0x50, 0x31, 0x7b, 0xf5, 0x49,
	0xbb, 0xbe, 0xab, 0xd7, 0x7f, 0x03, 0x00, 0x00, 0xff, 0xff, 0x9e, 0x84, 0x6a, 0x34, 0x2a, 0x03,
	0x00, 0x00,
}

func (m *MsgTransfer) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgTransfer) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgTransfer) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Memo) > 0 {
		i -= len(m.Memo)
		copy(dAtA[i:], m.Memo)
		i = encodeVarintIbcLegacy(dAtA, i, uint64(len(m.Memo)))
		i--
		dAtA[i] = 0x52
	}
	{
		size, err := m.Fee.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintIbcLegacy(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x4a
	if len(m.Router) > 0 {
		i -= len(m.Router)
		copy(dAtA[i:], m.Router)
		i = encodeVarintIbcLegacy(dAtA, i, uint64(len(m.Router)))
		i--
		dAtA[i] = 0x42
	}
	if m.TimeoutTimestamp != 0 {
		i = encodeVarintIbcLegacy(dAtA, i, uint64(m.TimeoutTimestamp))
		i--
		dAtA[i] = 0x38
	}
	{
		size, err := m.TimeoutHeight.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintIbcLegacy(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x32
	if len(m.Receiver) > 0 {
		i -= len(m.Receiver)
		copy(dAtA[i:], m.Receiver)
		i = encodeVarintIbcLegacy(dAtA, i, uint64(len(m.Receiver)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Sender) > 0 {
		i -= len(m.Sender)
		copy(dAtA[i:], m.Sender)
		i = encodeVarintIbcLegacy(dAtA, i, uint64(len(m.Sender)))
		i--
		dAtA[i] = 0x22
	}
	{
		size, err := m.Token.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintIbcLegacy(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.SourceChannel) > 0 {
		i -= len(m.SourceChannel)
		copy(dAtA[i:], m.SourceChannel)
		i = encodeVarintIbcLegacy(dAtA, i, uint64(len(m.SourceChannel)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.SourcePort) > 0 {
		i -= len(m.SourcePort)
		copy(dAtA[i:], m.SourcePort)
		i = encodeVarintIbcLegacy(dAtA, i, uint64(len(m.SourcePort)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintIbcLegacy(dAtA []byte, offset int, v uint64) int {
	offset -= sovIbcLegacy(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgTransfer) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.SourcePort)
	if l > 0 {
		n += 1 + l + sovIbcLegacy(uint64(l))
	}
	l = len(m.SourceChannel)
	if l > 0 {
		n += 1 + l + sovIbcLegacy(uint64(l))
	}
	l = m.Token.Size()
	n += 1 + l + sovIbcLegacy(uint64(l))
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovIbcLegacy(uint64(l))
	}
	l = len(m.Receiver)
	if l > 0 {
		n += 1 + l + sovIbcLegacy(uint64(l))
	}
	l = m.TimeoutHeight.Size()
	n += 1 + l + sovIbcLegacy(uint64(l))
	if m.TimeoutTimestamp != 0 {
		n += 1 + sovIbcLegacy(uint64(m.TimeoutTimestamp))
	}
	l = len(m.Router)
	if l > 0 {
		n += 1 + l + sovIbcLegacy(uint64(l))
	}
	l = m.Fee.Size()
	n += 1 + l + sovIbcLegacy(uint64(l))
	l = len(m.Memo)
	if l > 0 {
		n += 1 + l + sovIbcLegacy(uint64(l))
	}
	return n
}

func sovIbcLegacy(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozIbcLegacy(x uint64) (n int) {
	return sovIbcLegacy(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgTransfer) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowIbcLegacy
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
			return fmt.Errorf("proto: MsgTransfer: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgTransfer: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SourcePort", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIbcLegacy
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
				return ErrInvalidLengthIbcLegacy
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIbcLegacy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SourcePort = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SourceChannel", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIbcLegacy
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
				return ErrInvalidLengthIbcLegacy
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIbcLegacy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SourceChannel = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Token", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIbcLegacy
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
				return ErrInvalidLengthIbcLegacy
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthIbcLegacy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Token.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIbcLegacy
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
				return ErrInvalidLengthIbcLegacy
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIbcLegacy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Receiver", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIbcLegacy
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
				return ErrInvalidLengthIbcLegacy
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIbcLegacy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Receiver = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TimeoutHeight", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIbcLegacy
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
				return ErrInvalidLengthIbcLegacy
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthIbcLegacy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.TimeoutHeight.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TimeoutTimestamp", wireType)
			}
			m.TimeoutTimestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIbcLegacy
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TimeoutTimestamp |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Router", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIbcLegacy
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
				return ErrInvalidLengthIbcLegacy
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIbcLegacy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Router = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Fee", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIbcLegacy
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
				return ErrInvalidLengthIbcLegacy
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthIbcLegacy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Fee.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Memo", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIbcLegacy
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
				return ErrInvalidLengthIbcLegacy
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIbcLegacy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Memo = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipIbcLegacy(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthIbcLegacy
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
func skipIbcLegacy(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowIbcLegacy
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
					return 0, ErrIntOverflowIbcLegacy
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
					return 0, ErrIntOverflowIbcLegacy
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
				return 0, ErrInvalidLengthIbcLegacy
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupIbcLegacy
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthIbcLegacy
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthIbcLegacy        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowIbcLegacy          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupIbcLegacy = fmt.Errorf("proto: unexpected end of group")
)
