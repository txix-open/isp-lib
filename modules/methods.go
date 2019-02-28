package modules

const (
	userServiceGetCloudIdList     = "user/user/get_cloud_id_list"
	userServiceGetPhoneNumberList = "user/user/get_phone_list"
	userServiceGetEmailList       = "user/user/get_email_list"

	smsServiceSendSms = "sms/sms/send_sms_to_number"

	configServiceUpdateRemoteConfig = "config/config/update_active_config_by_instance_uuid_and_module_name"

	mdmGetRecord           = "mdm/data/get_record_by_external_id"
	mdmGetRecordList       = "mdm/data/get_records_list"
	mdmHandleRecord        = "mdm/data/handle_record_update"
	mdmHandleRecordBatch   = "mdm/data/handle_record_update_batch"
	mdmGetAttributes       = "mdm/attribute/get_all"
	mdmGetByExternalIdList = "mdm/data/get_by_external_id_list"

	mdmNotificationBroadcast = "mdm-notifier/notification/send"

	mdmConverterConvertToSudir               = "mdm-converter/sudir/convert"
	mdmConverterConvertToSudirBatch          = "mdm-converter/sudir/convert_batch"
	mdmConverterConvertToSudirBatchList      = "mdm-converter/sudir/convert_batch_list"
	mdmConverterConvertSearchRequestForSudir = "mdm-converter/sudir/convert_search_request"

	mdmConverterConvertToFind               = "mdm-converter/find/convert"
	mdmConverterConvertToFindBatch          = "mdm-converter/find/convert_batch"
	mdmConverterConvertToFindBatchList      = "mdm-converter/find/convert_batch_list"
	mdmConverterConvertSearchRequestForFind = "mdm-converter/find/convert_search_request"

	mdmConverterConvertAny          = "mdm-converter/any/convert"
	mdmConverterConvertAnyBatch     = "mdm-converter/any/convert_batch"
	mdmConverterConvertAnyBatchList = "mdm-converter/any/convert_batch_list"

	mdmConverterConvertErl          = "mdm-converter/erl/convert"
	mdmConverterConvertErlBatch     = "mdm-converter/erl/convert_batch"
	mdmConverterConvertErlBatchList = "mdm-converter/erl/convert_batch_list"

	mdmConverterFilterData          = "mdm-converter/data/filter"
	mdmConverterFilterDataBatch     = "mdm-converter/data/filter_batch"
	mdmConverterFilterDataBatchList = "mdm-converter/data/filter_batch_list"

	mdmConverterFilterAttributes = "mdm-converter/attribute/filter"

	mdmSearchSearchRecords      = "mdm-search/record/search"
	mdmSearchSearchIdList       = "mdm-search/record/search_id_list"
	mdmSearchSearchIdWithScroll = "mdm-search/record/search_id_with_scroll"

	mdtSearchPreferredSlices = "mdm-search/manage/preferred_slices_count"
)

var (
	MobilePushLinks = MobilePushServiceLinks{
		UserService: struct{ GetCouldIdList string }{GetCouldIdList: userServiceGetCloudIdList},
	}
	SmsLinks = SmsServiceLinks{
		UserService: struct{ GetPhoneNumberList string }{GetPhoneNumberList: userServiceGetPhoneNumberList},
	}
	EmailLinks = EmailServiceLinks{
		UserService: struct{ GetEmailList string }{GetEmailList: userServiceGetEmailList},
	}
	UserLinks = UserServiceLinks{
		SmsService: struct{ SendSms string }{SendSms: smsServiceSendSms},
	}
	MdmAdapterLinks = MdmAdapterServiceLinks{
		MdmService: struct {
			HandleRecord      string
			HandleRecordBatch string
		}{HandleRecord: mdmHandleRecord, HandleRecordBatch: mdmHandleRecordBatch},
	}
	MdmLinks = MdmServiceLinks{
		MdmNotifierService: struct{ BroadcastNotification string }{BroadcastNotification: mdmNotificationBroadcast},
	}
	MdmApiLinks = MdmApiServiceLinks{
		MdmService: struct {
			GetRecord           string
			GetAttributes       string
			GetRecordsList      string
			GetByExternalIdList string
		}{
			GetRecord:           mdmGetRecord,
			GetAttributes:       mdmGetAttributes,
			GetRecordsList:      mdmGetRecordList,
			GetByExternalIdList: mdmGetByExternalIdList,
		},
		MdmConverterService: struct {
			ConvertToSudir               string
			ConvertToFind                string
			ConvertSearchRequestForSudir string
			ConvertSearchRequestForFind  string

			FilterAttributes string
		}{
			ConvertToSudir:               mdmConverterConvertToSudir,
			ConvertToFind:                mdmConverterConvertToFind,
			ConvertSearchRequestForSudir: mdmConverterConvertSearchRequestForSudir,
			ConvertSearchRequestForFind:  mdmConverterConvertSearchRequestForFind,

			FilterAttributes: mdmConverterFilterAttributes,
		},
		MdmSearchService: struct {
			Search       string
			SearchIdList string
		}{
			Search:       mdmSearchSearchRecords,
			SearchIdList: mdmSearchSearchIdList,
		},
	}
	MdmNotifierLinks = MdmNotifyLinks{
		MdmConverterService: struct {
			ConvertToSudirBatchList string
			ConvertToFindBatchList  string
			ConvertAnyBatchList     string
			ConvertErlBatchList     string
		}{
			ConvertToSudirBatchList: mdmConverterConvertToSudirBatchList,
			ConvertToFindBatchList:  mdmConverterConvertToFindBatchList,
			ConvertAnyBatchList:     mdmConverterConvertAnyBatchList,
			ConvertErlBatchList:     mdmConverterConvertErlBatchList,
		},
	}
	MdmDumperLinks = MdmDumperService{
		MdmConverterService: struct {
			ConvertToSudirBatchList string
			ConvertToFindBatchList  string
			ConvertAnyBatchList     string
			ConvertErlBatchList     string
			FilterBatchList         string
		}{
			ConvertToSudirBatchList: mdmConverterConvertToSudirBatchList,
			ConvertToFindBatchList:  mdmConverterConvertToFindBatchList,
			ConvertAnyBatchList:     mdmConverterConvertAnyBatchList,
			ConvertErlBatchList:     mdmConverterConvertErlBatchList,
			FilterBatchList:         mdmConverterFilterDataBatchList,
		},
	}
	MdmAsyncApiLinks = MdmAsyncApiService{
		MdmConverterService: struct {
			ConvertToSudirBatchList string
			ConvertToFindBatchList  string
			ConvertAnyBatchList     string
			ConvertErlBatchList     string
			FilterBatchList         string
		}{
			ConvertToSudirBatchList: mdmConverterConvertToSudirBatchList,
			ConvertToFindBatchList:  mdmConverterConvertToFindBatchList,
			ConvertAnyBatchList:     mdmConverterConvertAnyBatchList,
			ConvertErlBatchList:     mdmConverterConvertErlBatchList,
			FilterBatchList:         mdmConverterFilterDataBatchList,
		},
		MdmSearchService: struct {
			SearchIdWithScroll   string
			PreferredSlicesCount string
		}{
			SearchIdWithScroll:   mdmSearchSearchIdWithScroll,
			PreferredSlicesCount: mdtSearchPreferredSlices,
		},
	}
	BackendLinks = AnyBackendServiceLinks{
		ConfigService: struct{ UpdateRemoteConfig string }{UpdateRemoteConfig: configServiceUpdateRemoteConfig},
	}
)
