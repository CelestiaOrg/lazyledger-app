// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: celestia/mint/v1/mint.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/gogo/protobuf/proto"
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

// Minter represents the minting state.
type Minter struct {
	// InflationRate is the rate at which new tokens should be minted for the
	// current year. For example if InflationRate=0.1, then 10% of the total
	// supply will be minted over the course of the year.
	InflationRate github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,1,opt,name=inflation_rate,json=inflationRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"inflation_rate"`
	// AnnualProvisions is the total number of tokens to be minted in the current
	// year due to inflation.
	AnnualProvisions github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,2,opt,name=annual_provisions,json=annualProvisions,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"annual_provisions"`
}

func (m *Minter) Reset()         { *m = Minter{} }
func (m *Minter) String() string { return proto.CompactTextString(m) }
func (*Minter) ProtoMessage()    {}
func (*Minter) Descriptor() ([]byte, []int) {
	return fileDescriptor_962d7cf1c9c59571, []int{0}
}
func (m *Minter) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Minter) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Minter.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Minter) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Minter.Merge(m, src)
}
func (m *Minter) XXX_Size() int {
	return m.Size()
}
func (m *Minter) XXX_DiscardUnknown() {
	xxx_messageInfo_Minter.DiscardUnknown(m)
}

var xxx_messageInfo_Minter proto.InternalMessageInfo

// Params holds parameters for the mint module.
type Params struct {
}

func (m *Params) Reset()      { *m = Params{} }
func (*Params) ProtoMessage() {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_962d7cf1c9c59571, []int{1}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Minter)(nil), "celestia.mint.v1.Minter")
	proto.RegisterType((*Params)(nil), "celestia.mint.v1.Params")
}

func init() { proto.RegisterFile("celestia/mint/v1/mint.proto", fileDescriptor_962d7cf1c9c59571) }

var fileDescriptor_962d7cf1c9c59571 = []byte{
	// 278 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4e, 0x4e, 0xcd, 0x49,
	0x2d, 0x2e, 0xc9, 0x4c, 0xd4, 0xcf, 0xcd, 0xcc, 0x2b, 0xd1, 0x2f, 0x33, 0x04, 0xd3, 0x7a, 0x05,
	0x45, 0xf9, 0x25, 0xf9, 0x42, 0x02, 0x30, 0x49, 0x3d, 0xb0, 0x60, 0x99, 0xa1, 0x94, 0x48, 0x7a,
	0x7e, 0x7a, 0x3e, 0x58, 0x52, 0x1f, 0xc4, 0x82, 0xa8, 0x93, 0x92, 0x4c, 0xce, 0x2f, 0xce, 0xcd,
	0x2f, 0x8e, 0x87, 0x48, 0x40, 0x38, 0x10, 0x29, 0xa5, 0x1b, 0x8c, 0x5c, 0x6c, 0xbe, 0x99, 0x79,
	0x25, 0xa9, 0x45, 0x42, 0xc9, 0x5c, 0x7c, 0x99, 0x79, 0x69, 0x39, 0x89, 0x25, 0x99, 0xf9, 0x79,
	0xf1, 0x45, 0x89, 0x25, 0xa9, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x4e, 0x36, 0x27, 0xee, 0xc9,
	0x33, 0xdc, 0xba, 0x27, 0xaf, 0x96, 0x9e, 0x59, 0x92, 0x51, 0x9a, 0xa4, 0x97, 0x9c, 0x9f, 0x0b,
	0x35, 0x03, 0x4a, 0xe9, 0x16, 0xa7, 0x64, 0xeb, 0x97, 0x54, 0x16, 0xa4, 0x16, 0xeb, 0xb9, 0xa4,
	0x26, 0x5f, 0xda, 0xa2, 0xcb, 0x05, 0xb5, 0xc2, 0x25, 0x35, 0x39, 0x88, 0x17, 0x6e, 0x66, 0x50,
	0x62, 0x49, 0xaa, 0x50, 0x26, 0x97, 0x60, 0x62, 0x5e, 0x5e, 0x69, 0x62, 0x0e, 0xc8, 0x31, 0x65,
	0x99, 0xc5, 0x99, 0xf9, 0x79, 0xc5, 0x12, 0x4c, 0x54, 0xb0, 0x47, 0x00, 0x62, 0x6c, 0x00, 0xdc,
	0x54, 0x25, 0x3e, 0x2e, 0xb6, 0x80, 0xc4, 0xa2, 0xc4, 0xdc, 0x62, 0x2b, 0x96, 0x19, 0x0b, 0xe4,
	0x19, 0x9c, 0xbc, 0x4e, 0x3c, 0x92, 0x63, 0xbc, 0xf0, 0x48, 0x8e, 0xf1, 0xc1, 0x23, 0x39, 0xc6,
	0x09, 0x8f, 0xe5, 0x18, 0x2e, 0x3c, 0x96, 0x63, 0xb8, 0xf1, 0x58, 0x8e, 0x21, 0xca, 0x00, 0xd9,
	0x46, 0x68, 0x90, 0xe6, 0x17, 0xa5, 0xc3, 0xd9, 0xba, 0x89, 0x05, 0x05, 0xfa, 0x15, 0x90, 0x18,
	0x00, 0xdb, 0x9f, 0xc4, 0x06, 0x0e, 0x3d, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0xc4, 0x50,
	0xa3, 0x7c, 0x9f, 0x01, 0x00, 0x00,
}

func (m *Minter) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Minter) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Minter) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.AnnualProvisions.Size()
		i -= size
		if _, err := m.AnnualProvisions.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintMint(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	{
		size := m.InflationRate.Size()
		i -= size
		if _, err := m.InflationRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintMint(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintMint(dAtA []byte, offset int, v uint64) int {
	offset -= sovMint(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Minter) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.InflationRate.Size()
	n += 1 + l + sovMint(uint64(l))
	l = m.AnnualProvisions.Size()
	n += 1 + l + sovMint(uint64(l))
	return n
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovMint(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMint(x uint64) (n int) {
	return sovMint(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Minter) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMint
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
			return fmt.Errorf("proto: Minter: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Minter: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InflationRate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
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
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.InflationRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AnnualProvisions", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
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
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.AnnualProvisions.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMint(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMint
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
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMint
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
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipMint(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMint
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
func skipMint(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMint
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
					return 0, ErrIntOverflowMint
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
					return 0, ErrIntOverflowMint
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
				return 0, ErrInvalidLengthMint
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMint
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMint
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMint        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMint          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMint = fmt.Errorf("proto: unexpected end of group")
)
