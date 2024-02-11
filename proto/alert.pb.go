// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.1
// source: proto/alert.proto

package proto

import (
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

type Metric_MetricType int32

const (
	Metric_UNSPECIFIED Metric_MetricType = 0
	Metric_COUNTER     Metric_MetricType = 1
	Metric_GAUGE       Metric_MetricType = 2
)

// Enum value maps for Metric_MetricType.
var (
	Metric_MetricType_name = map[int32]string{
		0: "UNSPECIFIED",
		1: "COUNTER",
		2: "GAUGE",
	}
	Metric_MetricType_value = map[string]int32{
		"UNSPECIFIED": 0,
		"COUNTER":     1,
		"GAUGE":       2,
	}
)

func (x Metric_MetricType) Enum() *Metric_MetricType {
	p := new(Metric_MetricType)
	*p = x
	return p
}

func (x Metric_MetricType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Metric_MetricType) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_alert_proto_enumTypes[0].Descriptor()
}

func (Metric_MetricType) Type() protoreflect.EnumType {
	return &file_proto_alert_proto_enumTypes[0]
}

func (x Metric_MetricType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Metric_MetricType.Descriptor instead.
func (Metric_MetricType) EnumDescriptor() ([]byte, []int) {
	return file_proto_alert_proto_rawDescGZIP(), []int{0, 0}
}

type Metric struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    string            `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Type  Metric_MetricType `protobuf:"varint,2,opt,name=type,proto3,enum=alert.Metric_MetricType" json:"type,omitempty"`
	Delta *int64            `protobuf:"varint,3,opt,name=delta,proto3,oneof" json:"delta,omitempty"`
	Value *float64          `protobuf:"fixed64,4,opt,name=value,proto3,oneof" json:"value,omitempty"`
}

func (x *Metric) Reset() {
	*x = Metric{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_alert_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Metric) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metric) ProtoMessage() {}

func (x *Metric) ProtoReflect() protoreflect.Message {
	mi := &file_proto_alert_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metric.ProtoReflect.Descriptor instead.
func (*Metric) Descriptor() ([]byte, []int) {
	return file_proto_alert_proto_rawDescGZIP(), []int{0}
}

func (x *Metric) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Metric) GetType() Metric_MetricType {
	if x != nil {
		return x.Type
	}
	return Metric_UNSPECIFIED
}

func (x *Metric) GetDelta() int64 {
	if x != nil && x.Delta != nil {
		return *x.Delta
	}
	return 0
}

func (x *Metric) GetValue() float64 {
	if x != nil && x.Value != nil {
		return *x.Value
	}
	return 0
}

type AddMetricRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metric *Metric `protobuf:"bytes,1,opt,name=metric,proto3" json:"metric,omitempty"`
}

func (x *AddMetricRequest) Reset() {
	*x = AddMetricRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_alert_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddMetricRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddMetricRequest) ProtoMessage() {}

func (x *AddMetricRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_alert_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddMetricRequest.ProtoReflect.Descriptor instead.
func (*AddMetricRequest) Descriptor() ([]byte, []int) {
	return file_proto_alert_proto_rawDescGZIP(), []int{1}
}

func (x *AddMetricRequest) GetMetric() *Metric {
	if x != nil {
		return x.Metric
	}
	return nil
}

type AddMetricResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metric *Metric `protobuf:"bytes,1,opt,name=metric,proto3" json:"metric,omitempty"`
}

func (x *AddMetricResponse) Reset() {
	*x = AddMetricResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_alert_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddMetricResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddMetricResponse) ProtoMessage() {}

func (x *AddMetricResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_alert_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddMetricResponse.ProtoReflect.Descriptor instead.
func (*AddMetricResponse) Descriptor() ([]byte, []int) {
	return file_proto_alert_proto_rawDescGZIP(), []int{2}
}

func (x *AddMetricResponse) GetMetric() *Metric {
	if x != nil {
		return x.Metric
	}
	return nil
}

var File_proto_alert_proto protoreflect.FileDescriptor

var file_proto_alert_proto_rawDesc = []byte{
	0x0a, 0x11, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x61, 0x6c, 0x65, 0x72, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x05, 0x61, 0x6c, 0x65, 0x72, 0x74, 0x22, 0xc7, 0x01, 0x0a, 0x06, 0x4d,
	0x65, 0x74, 0x72, 0x69, 0x63, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x2c, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x18, 0x2e, 0x61, 0x6c, 0x65, 0x72, 0x74, 0x2e, 0x4d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x12, 0x19, 0x0a, 0x05, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x03, 0x48, 0x00, 0x52, 0x05, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x88, 0x01, 0x01, 0x12, 0x19,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x01, 0x48, 0x01, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x88, 0x01, 0x01, 0x22, 0x35, 0x0a, 0x0a, 0x4d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0f, 0x0a, 0x0b, 0x55, 0x4e, 0x53, 0x50, 0x45,
	0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x43, 0x4f, 0x55, 0x4e,
	0x54, 0x45, 0x52, 0x10, 0x01, 0x12, 0x09, 0x0a, 0x05, 0x47, 0x41, 0x55, 0x47, 0x45, 0x10, 0x02,
	0x42, 0x08, 0x0a, 0x06, 0x5f, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x22, 0x39, 0x0a, 0x10, 0x41, 0x64, 0x64, 0x4d, 0x65, 0x74, 0x72, 0x69,
	0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x25, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x61, 0x6c, 0x65, 0x72, 0x74,
	0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x22,
	0x3a, 0x0a, 0x11, 0x41, 0x64, 0x64, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x25, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x61, 0x6c, 0x65, 0x72, 0x74, 0x2e, 0x4d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x32, 0x49, 0x0a, 0x07, 0x4d,
	0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x3e, 0x0a, 0x09, 0x41, 0x64, 0x64, 0x4d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x12, 0x17, 0x2e, 0x61, 0x6c, 0x65, 0x72, 0x74, 0x2e, 0x41, 0x64, 0x64, 0x4d,
	0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x61,
	0x6c, 0x65, 0x72, 0x74, 0x2e, 0x41, 0x64, 0x64, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x15, 0x5a, 0x13, 0x61, 0x6c, 0x65, 0x72, 0x74, 0x2d,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_alert_proto_rawDescOnce sync.Once
	file_proto_alert_proto_rawDescData = file_proto_alert_proto_rawDesc
)

func file_proto_alert_proto_rawDescGZIP() []byte {
	file_proto_alert_proto_rawDescOnce.Do(func() {
		file_proto_alert_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_alert_proto_rawDescData)
	})
	return file_proto_alert_proto_rawDescData
}

var file_proto_alert_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_alert_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_proto_alert_proto_goTypes = []interface{}{
	(Metric_MetricType)(0),    // 0: alert.Metric.MetricType
	(*Metric)(nil),            // 1: alert.Metric
	(*AddMetricRequest)(nil),  // 2: alert.AddMetricRequest
	(*AddMetricResponse)(nil), // 3: alert.AddMetricResponse
}
var file_proto_alert_proto_depIdxs = []int32{
	0, // 0: alert.Metric.type:type_name -> alert.Metric.MetricType
	1, // 1: alert.AddMetricRequest.metric:type_name -> alert.Metric
	1, // 2: alert.AddMetricResponse.metric:type_name -> alert.Metric
	2, // 3: alert.Metrics.AddMetric:input_type -> alert.AddMetricRequest
	3, // 4: alert.Metrics.AddMetric:output_type -> alert.AddMetricResponse
	4, // [4:5] is the sub-list for method output_type
	3, // [3:4] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_proto_alert_proto_init() }
func file_proto_alert_proto_init() {
	if File_proto_alert_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_alert_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Metric); i {
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
		file_proto_alert_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddMetricRequest); i {
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
		file_proto_alert_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddMetricResponse); i {
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
	file_proto_alert_proto_msgTypes[0].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_alert_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_alert_proto_goTypes,
		DependencyIndexes: file_proto_alert_proto_depIdxs,
		EnumInfos:         file_proto_alert_proto_enumTypes,
		MessageInfos:      file_proto_alert_proto_msgTypes,
	}.Build()
	File_proto_alert_proto = out.File
	file_proto_alert_proto_rawDesc = nil
	file_proto_alert_proto_goTypes = nil
	file_proto_alert_proto_depIdxs = nil
}