package modules

const (
	userServiceGetCloudIdList     = "user/user/get_cloud_id_list"
	userServiceGetPhoneNumberList = "user/user/get_phone_list"
	userServiceGetEmailList       = "user/user/get_email_list"

	smsServiceSendSms = "sms/sms/send_sms_to_number"

	configServiceUpdateRemoteConfig = "config/config/update_active_config_by_instance_uuid_and_module_name"

	mdmGetRecord         = "mdm/data/get_record_by_external_id"
	mdmHandleRecord      = "mdm/data/handle_record_update"
	mdmHandleRecordBatch = "mdm/data/handle_record_update_batch"

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

	mdmSearchSearchRecords = "mdm-search/record/search"
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
		MdmService: struct{ GetRecord string }{GetRecord: mdmGetRecord},
		MdmConverterService: struct {
			ConvertToSudir               string
			ConvertToFind                string
			ConvertSearchRequestForSudir string
			ConvertSearchRequestForFind  string
		}{
			ConvertToSudir:               mdmConverterConvertToSudir,
			ConvertToFind:                mdmConverterConvertToFind,
			ConvertSearchRequestForSudir: mdmConverterConvertSearchRequestForSudir,
			ConvertSearchRequestForFind:  mdmConverterConvertSearchRequestForFind,
		},
		MdmSearchService: struct{ Search string }{Search: mdmSearchSearchRecords},
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
	BackendLinks = AnyBackendServiceLinks{
		ConfigService: struct{ UpdateRemoteConfig string }{UpdateRemoteConfig: configServiceUpdateRemoteConfig},
	}
)
