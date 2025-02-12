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

// BlobHeader contains the information describing a blob and the way it is to be dispersed.
type BlobHeader struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The blob version. Blob versions are pushed onchain by EigenDA governance in an append only fashion and store the
	// maximum number of operators, number of chunks, and coding rate for a blob. On blob verification, these values
	// are checked against supplied or default security thresholds to validate the security assumptions of the
	// blob's availability.
	Version uint32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	// quorum_numbers is the list of quorum numbers that the blob is part of.
	// Each quorum will store the data, hence adding quorum numbers adds redundancy, making the blob more likely to be retrievable. Each quorum requires separate payment.
	//
	// On-demand dispersal is currently limited to using a subset of the following quorums:
	// - 0: ETH
	// - 1: EIGEN
	//
	// Reserved-bandwidth dispersal is free to use multiple quorums, however those must be reserved ahead of time. The quorum_numbers specified here must be a subset of the ones allowed by the on-chain reservation.
	// Check the allowed quorum numbers by looking up reservation struct: https://github.com/Layr-Labs/eigenda/blob/1430d56258b4e814b388e497320fd76354bfb478/contracts/src/interfaces/IPaymentVault.sol#L10
	QuorumNumbers []uint32 `protobuf:"varint,2,rep,packed,name=quorum_numbers,json=quorumNumbers,proto3" json:"quorum_numbers,omitempty"`
	// commitment is the KZG commitment to the blob
	Commitment *common.BlobCommitment `protobuf:"bytes,3,opt,name=commitment,proto3" json:"commitment,omitempty"`
	// payment_header contains payment information for the blob
	PaymentHeader *PaymentHeader `protobuf:"bytes,4,opt,name=payment_header,json=paymentHeader,proto3" json:"payment_header,omitempty"`
	// salt is used to ensure that the dispersal request is intentionally unique. This is currently only useful for
	// reserved payments when the same blob is submitted multiple times within the same reservation period. On-demand
	// payments already have unique cumulative_payment values for intentionally unique dispersal requests.
	Salt uint32 `protobuf:"varint,5,opt,name=salt,proto3" json:"salt,omitempty"`
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

func (x *BlobHeader) GetPaymentHeader() *PaymentHeader {
	if x != nil {
		return x.PaymentHeader
	}
	return nil
}

func (x *BlobHeader) GetSalt() uint32 {
	if x != nil {
		return x.Salt
	}
	return 0
}

// BlobCertificate contains a full description of a blob and how it is dispersed. Part of the certificate
// is provided by the blob submitter (i.e. the blob header), and part is provided by the disperser (i.e. the relays).
// Validator nodes eventually sign the blob certificate once they are in custody of the required chunks
// (note that the signature is indirect; validators sign the hash of a Batch, which contains the blob certificate).
type BlobCertificate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// blob_header contains data about the blob.
	BlobHeader *BlobHeader `protobuf:"bytes,1,opt,name=blob_header,json=blobHeader,proto3" json:"blob_header,omitempty"`
	// signature is an ECDSA signature signed by the blob request signer's account ID over the BlobHeader's blobKey,
	// which is a keccak hash of the serialized BlobHeader, and used to verify against blob dispersal request's account ID
	Signature []byte `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	// relay_keys is the list of relay keys that are in custody of the blob.
	// The relays custodying the data are chosen by the Disperser to which the DisperseBlob request was submitted.
	// It needs to contain at least 1 relay number.
	// To retrieve a blob from the relay, one can find that relay's URL in the EigenDARelayRegistry contract:
	// https://github.com/Layr-Labs/eigenda/blob/master/contracts/src/core/EigenDARelayRegistry.sol
	RelayKeys []uint32 `protobuf:"varint,3,rep,packed,name=relay_keys,json=relayKeys,proto3" json:"relay_keys,omitempty"`
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

func (x *BlobCertificate) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *BlobCertificate) GetRelayKeys() []uint32 {
	if x != nil {
		return x.RelayKeys
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

// PaymentHeader contains payment information for a blob.
// At least one of reservation_period or cumulative_payment must be set, and reservation_period
// is always considered before cumulative_payment. If reservation_period is set but not valid,
// the server will reject the request and not proceed with dispersal. If reservation_period is not set
// and cumulative_payment is set but not valid, the server will reject the request and not proceed with dispersal.
// Once the server has accepted the payment header, a client cannot cancel or rollback the payment.
// Every dispersal request will be charged by a multiple of `minNumSymbols` field defined by the payment vault contract.
// If the request blob size is smaller or not a multiple of `minNumSymbols`, the server will charge the user for the next
// multiple of `minNumSymbols` (https://github.com/Layr-Labs/eigenda/blob/1430d56258b4e814b388e497320fd76354bfb478/contracts/src/payments/PaymentVaultStorage.sol#L9).
type PaymentHeader struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The account ID of the disperser client. This account ID is an eth wallet address of the user,
	// corresponding to the key used by the client to sign the BlobHeader.
	AccountId string `protobuf:"bytes,1,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"`
	// The reservation period of the dispersal request is used for rate-limiting the user's account against their dedicated
	// bandwidth. This method requires users to set up reservation accounts with EigenDA team, and the team will set up an
	// on-chain record of reserved bandwidth for the user for some period of time. The dispersal client's accountant will set
	// this value to the current timestamp divided by the on-chain configured reservation period interval, mapping each request
	// to a time-based window and is serialized and parsed as a uint32. The disperser server then validates that it matches
	// either the current or the previous period.
	//
	// Example Usage Flow:
	//  1. The user sets up a reservation with the EigenDA team, including throughput (symbolsPerSecond), startTimestamp,
	//     endTimestamp, and reservationPeriodInterval.
	//  2. When sending a dispersal request at time t, the client computes reservation_period = floor(t / reservationPeriodInterval).
	//  3. The request includes this reservation_period index. The disperser checks:
	//     - If the reservation is active (t >= startTimestamp and t < endTimestamp).
	//     - After rounding up to the nearest multiple of `minNumSymbols` defined by the payment vault contract, the user still has enough bandwidth capacity
	//     (hasn’t exceeded symbolsPerSecond * reservationPeriodInterval).
	//  4. Server always go ahead with recording the received request, and then categorize the scenarios
	//     - If the remaining bandwidth is sufficient for the request, the dispersal request proceeds.
	//     - If the remaining bandwidth is not enough for the request, server fills up the current bin and overflowing the extra to a future bin.
	//     - If the bandwidth has already been exhausted, the request is rejected.
	//  5. Once the dispersal request signature has been verified, the server will not roll back the payment or the usage records.
	//     Users should be aware of this when planning their usage. The dispersal client written by EigenDA team takes account of this.
	//  6. When the reservation ends or usage is exhausted, the client must wait for the next reservation period or switch to on-demand.
	ReservationPeriod uint32 `protobuf:"varint,2,opt,name=reservation_period,json=reservationPeriod,proto3" json:"reservation_period,omitempty"`
	// Cumulative payment is the total amount of tokens paid by the requesting account, including the current request.
	// This value is serialized as an uint256 and parsed as a big integer, and must match the user’s on-chain deposit limits
	// as well as the recorded payments for all previous requests. Because it is a cumulative (not incremental) total,
	// requests can arrive out of order and still unambiguously declare how much of the on-chain deposit can be deducted.
	//
	// Example Decision Flow:
	//  1. In the set up phase, the user must deposit tokens into the EigenDA PaymentVault contract. The payment vault contract
	//     specifies the minimum number of symbols charged per dispersal, the pricing per symbol, and the maximum global rate for
	//     on-demand dispersals. The user should calculate the amount of tokens they would like to deposit based on their usage.
	//     The first time a user make a request, server will immediate read the contract for the on-chain balance. When user runs
	//     out of on-chain balance, the server will reject the request and not proceed with dispersal. When a user top up on-chain,
	//     the server will only refresh every few minutes for the top-up to take effect.
	//  2. The disperser client accounts how many tokens they’ve already paid (previousCumPmt).
	//  3. They should calculate the payment by rounding up blob size to the nearest multiple of `minNumSymbols` defined by the
	//     payment vault contract, and calculate the incremental amount of tokens needed for the current request needs based on
	//     protocol defined pricing.
	//  4. They take the sum of previousCumPmt + new incremental payment and place it in the “cumulative_payment” field.
	//  5. The disperser checks this new cumulative total against on-chain deposits and prior records (largest previous payment
	//     and smallest later payment if exists).
	//  6. If the payment number is valid, the request is confirmed and disperser proceeds with dispersal; otherwise it’s rejected.
	CumulativePayment []byte `protobuf:"bytes,3,opt,name=cumulative_payment,json=cumulativePayment,proto3" json:"cumulative_payment,omitempty"`
}

func (x *PaymentHeader) Reset() {
	*x = PaymentHeader{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_v2_common_v2_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PaymentHeader) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PaymentHeader) ProtoMessage() {}

func (x *PaymentHeader) ProtoReflect() protoreflect.Message {
	mi := &file_common_v2_common_v2_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PaymentHeader.ProtoReflect.Descriptor instead.
func (*PaymentHeader) Descriptor() ([]byte, []int) {
	return file_common_v2_common_v2_proto_rawDescGZIP(), []int{4}
}

func (x *PaymentHeader) GetAccountId() string {
	if x != nil {
		return x.AccountId
	}
	return ""
}

func (x *PaymentHeader) GetReservationPeriod() uint32 {
	if x != nil {
		return x.ReservationPeriod
	}
	return 0
}

func (x *PaymentHeader) GetCumulativePayment() []byte {
	if x != nil {
		return x.CumulativePayment
	}
	return nil
}

var File_common_v2_common_v2_proto protoreflect.FileDescriptor

var file_common_v2_common_v2_proto_rawDesc = []byte{
	0x0a, 0x19, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x32, 0x2f, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x5f, 0x76, 0x32, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x1a, 0x13, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xda, 0x01, 0x0a, 0x0a,
	0x42, 0x6c, 0x6f, 0x62, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x12, 0x25, 0x0a, 0x0e, 0x71, 0x75, 0x6f, 0x72, 0x75, 0x6d, 0x5f, 0x6e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x0d, 0x71, 0x75,
	0x6f, 0x72, 0x75, 0x6d, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x12, 0x36, 0x0a, 0x0a, 0x63,
	0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x16, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x42, 0x6c, 0x6f, 0x62, 0x43, 0x6f, 0x6d,
	0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x0a, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x6d,
	0x65, 0x6e, 0x74, 0x12, 0x3f, 0x0a, 0x0e, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x68,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x2e, 0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x48,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x0d, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x48, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x61, 0x6c, 0x74, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x04, 0x73, 0x61, 0x6c, 0x74, 0x22, 0x86, 0x01, 0x0a, 0x0f, 0x42, 0x6c, 0x6f,
	0x62, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x12, 0x36, 0x0a, 0x0b,
	0x62, 0x6c, 0x6f, 0x62, 0x5f, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x15, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x2e, 0x42, 0x6c,
	0x6f, 0x62, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x0a, 0x62, 0x6c, 0x6f, 0x62, 0x48, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x5f, 0x6b, 0x65, 0x79, 0x73,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x09, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x4b, 0x65, 0x79,
	0x73, 0x22, 0x62, 0x0a, 0x0b, 0x42, 0x61, 0x74, 0x63, 0x68, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72,
	0x12, 0x1d, 0x0a, 0x0a, 0x62, 0x61, 0x74, 0x63, 0x68, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x62, 0x61, 0x74, 0x63, 0x68, 0x52, 0x6f, 0x6f, 0x74, 0x12,
	0x34, 0x0a, 0x16, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x5f, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x14, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x22, 0x80, 0x01, 0x0a, 0x05, 0x42, 0x61, 0x74, 0x63, 0x68, 0x12,
	0x2e, 0x0a, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x16, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x2e, 0x42, 0x61, 0x74, 0x63,
	0x68, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12,
	0x47, 0x0a, 0x11, 0x62, 0x6c, 0x6f, 0x62, 0x5f, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x32, 0x2e, 0x42, 0x6c, 0x6f, 0x62, 0x43, 0x65, 0x72, 0x74, 0x69,
	0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x52, 0x10, 0x62, 0x6c, 0x6f, 0x62, 0x43, 0x65, 0x72, 0x74,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x73, 0x22, 0x8c, 0x01, 0x0a, 0x0d, 0x50, 0x61, 0x79,
	0x6d, 0x65, 0x6e, 0x74, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x2d, 0x0a, 0x12, 0x72, 0x65, 0x73,
	0x65, 0x72, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x11, 0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x12, 0x2d, 0x0a, 0x12, 0x63, 0x75, 0x6d, 0x75,
	0x6c, 0x61, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x11, 0x63, 0x75, 0x6d, 0x75, 0x6c, 0x61, 0x74, 0x69, 0x76, 0x65,
	0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x4c, 0x61, 0x79, 0x72, 0x2d, 0x4c, 0x61, 0x62, 0x73, 0x2f,
	0x65, 0x69, 0x67, 0x65, 0x6e, 0x64, 0x61, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x72, 0x70, 0x63,
	0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x32, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
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

var file_common_v2_common_v2_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_common_v2_common_v2_proto_goTypes = []interface{}{
	(*BlobHeader)(nil),            // 0: common.v2.BlobHeader
	(*BlobCertificate)(nil),       // 1: common.v2.BlobCertificate
	(*BatchHeader)(nil),           // 2: common.v2.BatchHeader
	(*Batch)(nil),                 // 3: common.v2.Batch
	(*PaymentHeader)(nil),         // 4: common.v2.PaymentHeader
	(*common.BlobCommitment)(nil), // 5: common.BlobCommitment
}
var file_common_v2_common_v2_proto_depIdxs = []int32{
	5, // 0: common.v2.BlobHeader.commitment:type_name -> common.BlobCommitment
	4, // 1: common.v2.BlobHeader.payment_header:type_name -> common.v2.PaymentHeader
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
		file_common_v2_common_v2_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PaymentHeader); i {
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
			NumMessages:   5,
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
