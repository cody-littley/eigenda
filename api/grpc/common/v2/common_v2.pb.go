// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.23.4
// source: common/v2/common_v2.proto

package v2

import (
	common "github.com/Layr-Labs/eigenda/api/grpc/common"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// BlobHeader is the header of a blob
type BlobHeader struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Blob version
	Version uint32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	// quorum_numbers is the list of quorum numbers that the blob is part of
	QuorumNumbers []uint32 `protobuf:"varint,2,rep,packed,name=quorum_numbers,json=quorumNumbers,proto3" json:"quorum_numbers,omitempty"`
	// commitment is the KZG commitment of the blob
	Commitment *common.BlobCommitment `protobuf:"bytes,3,opt,name=commitment,proto3" json:"commitment,omitempty"`
	// payment_header contains payment information for the blob
	PaymentHeader *common.PaymentHeader `protobuf:"bytes,4,opt,name=payment_header,json=paymentHeader,proto3" json:"payment_header,omitempty"`
	// signature over keccak hash of the blob_header that can be verified by blob_header.account_id
	Signature []byte `protobuf:"bytes,5,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (x *BlobHeader) Reset() {
	*x = BlobHeader{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v2_common_v2_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlobHeader) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlobHeader) ProtoMessage() {}

func (x *BlobHeader) ProtoReflect() protoreflect.Message {
	mi := &file_common_v2_common_v2_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlobHeader.ProtoReflect.Descriptor instead.
func (*BlobHeader) Descriptor() ([]byte, []int) {
	return file_common_v2_common_v2_proto_rawDescGZIP(), []int{0}
}

func (x *BlobHeader) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *BlobHeader) GetQuorumNumbers() []uint32 {
	if x != nil {
		return x.QuorumNumbers
	}
	return nil
}

func (x *BlobHeader) GetCommitment() *common.BlobCommitment {
	if x != nil {
		return x.Commitment
	}
	return nil
}

func (x *BlobHeader) GetPaymentHeader() *common.PaymentHeader {
	if x != nil {
		return x.PaymentHeader
	}
	return nil
}

func (x *BlobHeader) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

// BlobCertificate is what gets attested by the network
type BlobCertificate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// blob_header is the header of the blob
	BlobHeader *BlobHeader `protobuf:"bytes,1,opt,name=blob_header,json=blobHeader,proto3" json:"blob_header,omitempty"`
	// relays is the list of relays that are in custody of the blob
	Relays []uint32 `protobuf:"varint,2,rep,packed,name=relays,proto3" json:"relays,omitempty"`
}

func (x *BlobCertificate) Reset() {
	*x = BlobCertificate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v2_common_v2_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlobCertificate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlobCertificate) ProtoMessage() {}

func (x *BlobCertificate) ProtoReflect() protoreflect.Message {
	mi := &file_common_v2_common_v2_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlobCertificate.ProtoReflect.Descriptor instead.
func (*BlobCertificate) Descriptor() ([]byte, []int) {
	return file_common_v2_common_v2_proto_rawDescGZIP(), []int{1}
}

func (x *BlobCertificate) GetBlobHeader() *BlobHeader {
	if x != nil {
		return x.BlobHeader
	}
	return nil
}

func (x *BlobCertificate) GetRelays() []uint32 {
	if x != nil {
		return x.Relays
	}
	return nil
}

// BatchHeader is the header of a batch of blobs
type BatchHeader struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// batch_root is the root of the merkle tree of the hashes of blob certificates in the batch
	BatchRoot []byte `protobuf:"bytes,1,opt,name=batch_root,json=batchRoot,proto3" json:"batch_root,omitempty"`
	// reference_block_number is the block number that the state of the batch is based on for attestation
	ReferenceBlockNumber uint64 `protobuf:"varint,2,opt,name=reference_block_number,json=referenceBlockNumber,proto3" json:"reference_block_number,omitempty"`
}

func (x *BatchHeader) Reset() {
	*x = BatchHeader{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v2_common_v2_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BatchHeader) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchHeader) ProtoMessage() {}

func (x *BatchHeader) ProtoReflect() protoreflect.Message {
	mi := &file_common_v2_common_v2_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchHeader.ProtoReflect.Descriptor instead.
func (*BatchHeader) Descriptor() ([]byte, []int) {
	return file_common_v2_common_v2_proto_rawDescGZIP(), []int{2}
}

func (x *BatchHeader) GetBatchRoot() []byte {
	if x != nil {
		return x.BatchRoot
	}
	return nil
}

func (x *BatchHeader) GetReferenceBlockNumber() uint64 {
	if x != nil {
		return x.ReferenceBlockNumber
	}
	return 0
}

// Batch is a batch of blob certificates
type Batch struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// header contains metadata about the batch
	Header *BatchHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	// blob_certificates is the list of blob certificates in the batch
	BlobCertificates []*BlobCertificate `protobuf:"bytes,2,rep,name=blob_certificates,json=blobCertificates,proto3" json:"blob_certificates,omitempty"`
}

func (x *Batch) Reset() {
	*x = Batch{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v2_common_v2_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Batch) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Batch) ProtoMessage() {}

func (x *Batch) ProtoReflect() protoreflect.Message {
	mi := &file_common_v2_common_v2_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Batch.ProtoReflect.Descriptor instead.
func (*Batch) Descriptor() ([]byte, []int) {
	return file_common_v2_common_v2_proto_rawDescGZIP(), []int{3}
}

func (x *Batch) GetHeader() *BatchHeader {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *Batch) GetBlobCertificates() []*BlobCertificate {
	if x != nil {
		return x.BlobCertificates
	}
	return nil
}

var File_common_v2_common_v2_proto protoreflect.FileDescriptor

var file_common_v2_common_v2_proto_rawDesc = []byte{
	0x0a, 0x19, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x32, 0x2f, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x5f, 0x76, 0x32, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x1a, 0x13, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xe1, 0x01, 0x0a, 0x0a,
	0x42, 0x6c, 0x6f, 0x62, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x12, 0x25, 0x0a, 0x0e, 0x71, 0x75, 0x6f, 0x72, 0x75, 0x6d, 0x5f, 0x6e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x0d, 0x71, 0x75,
	0x6f, 0x72, 0x75, 0x6d, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x12, 0x36, 0x0a, 0x0a, 0x63,
	0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x16, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x42, 0x6c, 0x6f, 0x62, 0x43, 0x6f, 0x6d,
	0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x0a, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x6d,
	0x65, 0x6e, 0x74, 0x12, 0x3c, 0x0a, 0x0e, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x68,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x48, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x52, 0x0d, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x48, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x22,
	0x61, 0x0a, 0x0f, 0x42, 0x6c, 0x6f, 0x62, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x65, 0x12, 0x36, 0x0a, 0x0b, 0x62, 0x6c, 0x6f, 0x62, 0x5f, 0x68, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2e, 0x76, 0x32, 0x2e, 0x42, 0x6c, 0x6f, 0x62, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x0a,
	0x62, 0x6c, 0x6f, 0x62, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65,
	0x6c, 0x61, 0x79, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x06, 0x72, 0x65, 0x6c, 0x61,
	0x79, 0x73, 0x22, 0x62, 0x0a, 0x0b, 0x42, 0x61, 0x74, 0x63, 0x68, 0x48, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x12, 0x1d, 0x0a, 0x0a, 0x62, 0x61, 0x74, 0x63, 0x68, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x62, 0x61, 0x74, 0x63, 0x68, 0x52, 0x6f, 0x6f, 0x74,
	0x12, 0x34, 0x0a, 0x16, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x5f, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x14, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b,
	0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x22, 0x80, 0x01, 0x0a, 0x05, 0x42, 0x61, 0x74, 0x63, 0x68,
	0x12, 0x2e, 0x0a, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x16, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x2e, 0x42, 0x61, 0x74,
	0x63, 0x68, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72,
	0x12, 0x47, 0x0a, 0x11, 0x62, 0x6c, 0x6f, 0x62, 0x5f, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69,
	0x63, 0x61, 0x74, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x2e, 0x42, 0x6c, 0x6f, 0x62, 0x43, 0x65, 0x72, 0x74,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x52, 0x10, 0x62, 0x6c, 0x6f, 0x62, 0x43, 0x65, 0x72,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x4c, 0x61, 0x79, 0x72, 0x2d, 0x4c, 0x61, 0x62,
	0x73, 0x2f, 0x65, 0x69, 0x67, 0x65, 0x6e, 0x64, 0x61, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x72,
	0x70, 0x63, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x32, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_common_v2_common_v2_proto_rawDescOnce sync.Once
	file_common_v2_common_v2_proto_rawDescData = file_common_v2_common_v2_proto_rawDesc
)

func file_common_v2_common_v2_proto_rawDescGZIP() []byte {
	file_common_v2_common_v2_proto_rawDescOnce.Do(func() {
		file_common_v2_common_v2_proto_rawDescData = protoimpl.X.CompressGZIP(file_common_v2_common_v2_proto_rawDescData)
	})
	return file_common_v2_common_v2_proto_rawDescData
}

var file_common_v2_common_v2_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_common_v2_common_v2_proto_goTypes = []interface{}{
	(*BlobHeader)(nil),            // 0: common.v2.BlobHeader
	(*BlobCertificate)(nil),       // 1: common.v2.BlobCertificate
	(*BatchHeader)(nil),           // 2: common.v2.BatchHeader
	(*Batch)(nil),                 // 3: common.v2.Batch
	(*common.BlobCommitment)(nil), // 4: common.BlobCommitment
	(*common.PaymentHeader)(nil),  // 5: common.PaymentHeader
}
var file_common_v2_common_v2_proto_depIdxs = []int32{
	4, // 0: common.v2.BlobHeader.commitment:type_name -> common.BlobCommitment
	5, // 1: common.v2.BlobHeader.payment_header:type_name -> common.PaymentHeader
	0, // 2: common.v2.BlobCertificate.blob_header:type_name -> common.v2.BlobHeader
	2, // 3: common.v2.Batch.header:type_name -> common.v2.BatchHeader
	1, // 4: common.v2.Batch.blob_certificates:type_name -> common.v2.BlobCertificate
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_common_v2_common_v2_proto_init() }
func file_common_v2_common_v2_proto_init() {
	if File_common_v2_common_v2_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_common_v2_common_v2_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlobHeader); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_v2_common_v2_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlobCertificate); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_v2_common_v2_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BatchHeader); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_common_v2_common_v2_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Batch); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_common_v2_common_v2_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_v2_common_v2_proto_goTypes,
		DependencyIndexes: file_common_v2_common_v2_proto_depIdxs,
		MessageInfos:      file_common_v2_common_v2_proto_msgTypes,
	}.Build()
	File_common_v2_common_v2_proto = out.File
	file_common_v2_common_v2_proto_rawDesc = nil
	file_common_v2_common_v2_proto_goTypes = nil
	file_common_v2_common_v2_proto_depIdxs = nil
}