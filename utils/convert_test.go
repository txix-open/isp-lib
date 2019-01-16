package utils

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/golang/protobuf/ptypes/struct"
	"github.com/json-iterator/go"
	"sync"
	"testing"
)

var (
	srcJson  = []byte(`{"users": [{"enabled": true, "id_type": "dolor pariatur ip", "id_value": "nisi laboris Duis", "etalon_id": "mollit sunt", "citizen_id": "Duis ipsum enim deserunt", "ext_upd_date": "2015-01-22T12:14:19.576Z", "last_update_dt": "1974-03-17T15:09:28.305Z", "is_confirmed_offline": true}, {"enabled": true, "id_type": "ad dolore tempor non", "id_value": "qui nulla", "etalon_id": "aute", "citizen_id": "in", "ext_upd_date": "1944-05-04T05:20:11.755Z", "last_update_dt": "2015-05-06T13:51:59.582Z", "is_confirmed_offline": true}, {"enabled": false, "id_type": "in mollit", "id_value": "dolore cupidatat labori", "etalon_id": "eiusmod enim", "citizen_id": "nulla nisi aliqua", "ext_upd_date": "1988-05-12T02:00:03.962Z", "last_update_dt": "1979-05-01T07:40:59.072Z", "is_confirmed_offline": false}], "gkuPays": [{"epd": "ut", "phone": "cillum laborum minim ut", "mesmeter": "consequat in ex magna amet", "etalon_id": "in adipisicing labore", "mesaccount": "in Ut dolor officia consequat", "epd_comment": "dolore", "chronicle_id": "nulla", "last_update_dt": "2006-11-02T04:02:29.618Z", "entity_active_ind": false, "gkupay_address_rel": "laboris tempor minim ad"}, {"epd": "id cillum deserunt elit", "phone": "eu", "mesmeter": "minim deserunt irure", "etalon_id": "tempor amet in eiusmod elit", "mesaccount": "in Lorem et", "epd_comment": "consequat dolore", "chronicle_id": "enim aliqua id ut", "last_update_dt": "1941-11-29T23:42:02.182Z", "entity_active_ind": false, "gkupay_address_rel": "velit fugiat"}, {"epd": "Duis eu", "phone": "aute reprehenderit dolore", "mesmeter": "dolor irure occaecat", "etalon_id": "cillum", "mesaccount": "minim consequat", "epd_comment": "dolore", "chronicle_id": "ut", "last_update_dt": "2015-12-01T02:26:05.524Z", "entity_active_ind": false, "gkupay_address_rel": "consequat reprehenderit velit irure in"}], "pending": true, "birth_dt": "1971-06-09T14:15:39.458Z", "contacts": [{"ref_num": "culpa officia", "etalon_id": "sit", "external_id": "sunt sed", "last_update_dt": "1971-04-05T03:29:18.010Z", "cont_meth_cat_cd": "cupidatat Ut"}, {"ref_num": "ex cillum aliquip cupid", "etalon_id": "magna officia elit ipsum commodo", "external_id": "est dolore", "last_update_dt": "1987-09-13T22:39:08.892Z", "cont_meth_cat_cd": "tempor anim occaecat reprehenderit ea"}, {"ref_num": "Duis velit Ut", "etalon_id": "in culpa aliquip elit consecte", "external_id": "voluptate esse", "last_update_dt": "1968-09-12T06:19:39.633Z", "cont_meth_cat_cd": "pariatur do ad"}, {"ref_num": "occaecat est", "etalon_id": "Except", "external_id": "id", "last_update_dt": "1986-12-02T03:57:53.032Z", "cont_meth_cat_cd": "pariatur"}], "del_sign": false, "start_dt": "1973-11-21T21:40:44.754Z", "vehicles": [{"vin_num": "ullamco ipsum proident", "del_sign": true, "etalon_id": "dolore dolor in enim", "ptsnumber": "Ut", "stsnumber": "do ullamco officia", "reg_number": "in", "description": "incididunt mollit fugiat", "chronicle_id": "eiusmod consequat deserunt", "temporary_id": "minim incididunt", "last_update_dt": "1961-09-11T14:57:20.528Z", "vehicles_docs_rel": [{"document_id": "fugiat eiusmod commodo"}, {"document_id": "in deserunt ex in irure"}, {"document_id": "nisi tempor aliqua dolor laboris"}, {"document_id": "commodo ut ad"}, {"document_id": "ut"}]}, {"vin_num": "exercitation laboris dolor", "del_sign": false, "etalon_id": "proident et elit Lorem consectetur", "ptsnumber": "consequat ipsum", "stsnumber": "ex", "reg_number": "esse voluptate sit deserun", "description": "tempor dolore aute enim", "chronicle_id": "qui Excepteur dolor", "temporary_id": "ea ad sit", "last_update_dt": "1973-12-06T00:00:52.786Z", "vehicles_docs_rel": [{"document_id": "dolore elit adipisicing"}, {"document_id": "ea reprehenderit"}, {"document_id": "consectetur Lorem reprehenderit culpa"}, {"document_id": "anim enim"}, {"document_id": "enim eiusmod dolor"}]}, {"vin_num": "elit sit pariatur non aute", "del_sign": false, "etalon_id": "dolor", "ptsnumber": "id in nulla min", "stsnumber": "nulla", "reg_number": "et in elit", "description": "fugiat adipisicing deserun", "chronicle_id": "deserunt eiusmod", "temporary_id": "non velit incididunt", "last_update_dt": "2010-01-28T02:39:54.355Z", "vehicles_docs_rel": [{"document_id": "dolore cillum laborum qui"}, {"document_id": "nulla voluptate dolor"}, {"document_id": "magna"}, {"document_id": "culpa Excepteur deserunt amet"}, {"document_id": "culpa"}]}, {"vin_num": "in laboris consequat dolor", "del_sign": false, "etalon_id": "irure occaecat aute deserunt dolo", "ptsnumber": "ea", "stsnumber": "dolor aliquip dolor non d", "reg_number": "eu dolore aute sit incididunt", "description": "eu proident aliqua minim sit", "chronicle_id": "occaecat veniam nulla qui enim", "temporary_id": "sit Duis", "last_update_dt": "1980-01-09T16:45:33.434Z", "vehicles_docs_rel": [{"document_id": "sint tempor laboris laborum"}, {"document_id": "ad amet Duis sit"}, {"document_id": "magna"}, {"document_id": "cillum dolore"}]}, {"vin_num": "sint Lorem sed", "del_sign": false, "etalon_id": "dolore aliqua in consectetur", "ptsnumber": "aliqua pariatur", "stsnumber": "officia minim", "reg_number": "nostrud aliqua nulla", "description": "eiusmod", "chronicle_id": "ut", "temporary_id": "non consectetur est quis", "last_update_dt": "2006-06-04T15:23:44.060Z", "vehicles_docs_rel": [{"document_id": "ut dolore dolor"}, {"document_id": "sint occaecat dolor Excepteur ut"}, {"document_id": "minim aliqua"}, {"document_id": "esse"}]}], "addresses": [{"unad": "nostrud sed cupidatat cillum", "unom": "do qui", "end_dt": "1945-04-30T19:24:54.191Z", "region": "reprehenderit dolore elit Duis", "room_no": "commodo laboris", "house_no": "enim culpa laboris", "kladr_id": "ex in si", "city_name": "mollit incididun", "corpus_no": "officia pariatur nulla", "etalon_id": "velit ut ut nostrud occaecat", "street_id": "in laboris tempor proident commodo", "street_omk": "ullamco consequat voluptate", "external_id": "Excepteur", "postal_code": "est commodo elit", "stroenie_no": "Excepteur incididunt in amet", "temporary_id": "ea elit incididunt consequat Ut", "addr_line_one": "commodo ut velit nisi", "addr_line_two": "", "building_name": "ea occaecat laboris esse minim", "residence_num": "vel", "last_update_dt": "1959-05-17T06:52:29.082Z", "street_name_ex": "Excepteur", "addr_line_three": "pariatur in", "moscow_area_name": "est sint", "last_verified_date": "2016-02-03T05:43:08.400Z", "moscow_area_ext_id": "commodo do pariatur id", "moscow_district_id": "qui", "moscow_district_name": "veniam et minim dolore cillum"}, {"unad": "nostrud nisi", "unom": "ullamco", "end_dt": "1982-05-02T05:02:56.038Z", "region": "consequat", "room_no": "exercitation laborum cillum velit consequat", "house_no": "nostrud", "kladr_id": "officia consectetur eu Duis", "city_name": "non ullamco ut", "corpus_no": "ullamco", "etalon_id": "dolor dolore", "street_id": "anim cillum nisi ipsum", "street_omk": "consectetur dolor in", "external_id": "cupidatat", "postal_code": "incididunt in sunt", "stroenie_no": "nisi officia do", "temporary_id": "laborum eu cillum velit in", "addr_line_one": "fugiat minim velit", "addr_line_two": "sunt", "building_name": "reprehenderit ex", "residence_num": "sit magna", "last_update_dt": "1950-05-18T07:54:42.230Z", "street_name_ex": "est", "addr_line_three": "enim sit", "moscow_area_name": "amet irure non eiusmod sint", "last_verified_date": "2017-10-30T14:02:03.976Z", "moscow_area_ext_id": "est officia", "moscow_district_id": "anim", "moscow_district_name": "id aliqua adipisicing est"}, {"unad": "tempor reprehenderit velit eu ipsum", "unom": "quis ipsum do pariatur tempor", "end_dt": "1951-05-10T05:43:17.362Z", "region": "proident", "room_no": "Excepteur", "house_no": "aute cupidatat", "kladr_id": "non irure qui", "city_name": "veli", "corpus_no": "exercitation proident ad ut", "etalon_id": "deserun", "street_id": "consequat", "street_omk": "quis laborum amet in velit", "external_id": "eiusmod laboris nostrud", "postal_code": "aliqua Duis ipsum occaecat deserunt", "stroenie_no": "veniam sunt eiusmod tempor sint", "temporary_id": "consequat nostrud", "addr_line_one": "in", "addr_line_two": "occaecat in", "building_name": "et", "residence_num": "cupidatat enim minim commodo est", "last_update_dt": "1991-11-08T17:18:24.228Z", "street_name_ex": "laboris eu do sint", "addr_line_three": "occaeca", "moscow_area_name": "fugiat minim in aute magna", "last_verified_date": "1997-05-25T00:30:23.029Z", "moscow_area_ext_id": "culpa consequat cillum nisi nulla", "moscow_district_id": "sed quis proident", "moscow_district_name": "in reprehenderit Excepteur quis"}, {"unad": "non aute elit esse", "unom": "nulla exercitation", "end_dt": "1979-03-18T11:08:10.609Z", "region": "anim aliquip e", "room_no": "enim cillum dolore", "house_no": "ut i", "kladr_id": "Excepteur proident ", "city_name": "commodo", "corpus_no": "proident qui cillum incididunt", "etalon_id": "aute Excepteur", "street_id": "nostrud no", "street_omk": "nostrud nisi", "external_id": "officia", "postal_code": "qui cillum", "stroenie_no": "magna enim Excepteur irure sit", "temporary_id": "ad magna et ", "addr_line_one": "cupidatat", "addr_line_two": "fugiat laboris esse", "building_name": "adipisicing", "residence_num": "quis deserunt sit est sed", "last_update_dt": "2015-05-09T03:12:12.526Z", "street_name_ex": "ipsum", "addr_line_three": "reprehenderit", "moscow_area_name": "officia", "last_verified_date": "1967-01-16T09:09:42.802Z", "moscow_area_ext_id": "sit tempor", "moscow_district_id": "laborum", "moscow_district_name": "Excepteur ad sit adipisicing"}, {"unad": "cupidatat Lorem dolor consequat", "unom": "", "end_dt": "1939-08-15T10:29:02.512Z", "region": "non ullamco adipisicing sunt", "room_no": "elit enim culpa", "house_no": "ea dolor sunt quis eu", "kladr_id": "in aliqua non do", "city_name": "elit incididunt irure", "corpus_no": "consequat dolore adipisicing cillum exercitation", "etalon_id": "ut ex dolore in", "street_id": "Duis ex aliqua ea ad", "street_omk": "ullamco Excepteur aute", "external_id": "Dui", "postal_code": "veniam", "stroenie_no": "ve", "temporary_id": "commodo incididunt ad aliq", "addr_line_one": "deserunt eiusmod", "addr_line_two": "elit", "building_name": "culpa incididunt", "residence_num": "ea Excepteur irure non", "last_update_dt": "2017-06-23T14:28:38.881Z", "street_name_ex": "eiusmod Excepteu", "addr_line_three": "esse sunt", "moscow_area_name": "consequat id sunt tempor qui", "last_verified_date": "1975-05-21T06:15:30.800Z", "moscow_area_ext_id": "deserunt magna non", "moscow_district_id": "et nisi Lorem", "moscow_district_name": "labore amet"}], "documents": [{"code": "dolor laboris ut", "end_dt": "2014-01-03T16:09:02.841Z", "comment": "commodo", "ref_num": "eiusmod Duis in aliquip", "id_tp_cd": "#deph_cert", "start_dt": "1975-04-18T10:28:39.178Z", "etalon_id": "laborum ipsum minim cupida", "expiry_dt": "1960-10-20T21:18:21.551Z", "assigned_by": "anim do proident sunt eu", "ext_upd_date": "2018-05-18T20:25:36.785Z", "originblock ": "culpa id minim dolore voluptate", "temporary_id": "minim Ut esse Excepteur", "last_update_dt": "1965-04-13T20:47:30.927Z", "last_verified_dt": "1982-11-15T07:59:32.700Z", "last_update_tx_id": "nisi am", "identification_issuer": "fugiat tempor cillum aliquip nisi", "identification_series": "ut aute mollit eiu", "identification_issuer_code": "dolore", "identification_passport_birth_place": "qui eiusmod "}, {"code": "consectetur ut dolore", "end_dt": "2009-10-18T08:05:09.278Z", "comment": "cupidatat dolor ullamco nisi Excepteur", "ref_num": "tempor", "id_tp_cd": "#deph_cert", "start_dt": "1953-02-25T08:47:29.780Z", "etalon_id": "ullamco magna adipisicing", "expiry_dt": "1984-08-04T15:44:22.928Z", "assigned_by": "officia consectetur irure", "ext_upd_date": "1967-02-20T22:28:15.422Z", "originblock ": "veniam dolor", "temporary_id": "amet in eiusmod", "last_update_dt": "1965-10-26T23:36:22.317Z", "last_verified_dt": "1949-11-22T12:55:35.115Z", "last_update_tx_id": "amet anim aliquip non", "identification_issuer": "incididunt nulla esse proident", "identification_series": "cillum tempor", "identification_issuer_code": "sed v", "identification_passport_birth_place": "reprehenderit amet"}], "etalon_id": "7d0f75fb-112a-4585-9fde-bad3c2a30d51", "last_name": "dolore laboris voluptate", "spec_type": 55625668, "dirtystage": -42811660, "failedmask": 94714496, "ext_upd_date": "2004-11-14T08:50:21.924Z", "temporary_id": "magna ex esse velit", "date_of_death": "1995-09-20T06:05:41.476Z", "esCredentials": [{"login": "labore aliquip elit", "password": "dolore adipisicing aliqua laboris dolor", "etalon_id": "magna dolor", "description": "labore", "chronicle_id": "proident irure incididunt mollit amet", "escred_tp_cd": "dolore fugiat irure", "last_update_dt": "1996-04-28T00:23:55.413Z"}, {"login": "adipisicing sint", "password": "sed mollit tempor cupidatat", "etalon_id": "proident esse sint", "description": "pariatur elit ea", "chronicle_id": "ipsum Excepteur sit aliqua", "escred_tp_cd": "aliqua Excepteur ut", "last_update_dt": "1958-12-01T05:26:03.541Z"}, {"login": "labore", "password": "laboris dolor cupidatat deserunt", "etalon_id": "veniam", "description": "nulla conse", "chronicle_id": "veniam ipsum officia", "escred_tp_cd": "Ut sint dolor qui commodo", "last_update_dt": "1979-05-03T00:01:53.754Z"}, {"login": "culpa minim", "password": "laborum ut aute", "etalon_id": "et occaecat labore proident elit", "description": "adipisicing ad nisi", "chronicle_id": "cupidatat labore ipsum nisi", "escred_tp_cd": "fugiat occaecat", "last_update_dt": "1969-02-21T09:06:47.486Z"}], "has_privilege": false, "sourcechannel": 58528287, "gender_tp_code": "reprehenderit Duis enim", "given_name_one": "est Lorem dolore eiusmod", "given_name_two": "ex", "last_update_dt": "2002-06-28T01:05:44.574Z", "validationmask": -10206894, "citizen_relatives": [{"citizen": [{}, {}], "start_dt": "1978-07-30T17:30:08.858Z", "etalon_id": "aute ut consectetur", "rel_tp_cd": "Duis mollit elit", "access_level": 62036373, "last_update_dt": "1943-07-31T00:21:31.023Z"}, {"citizen": [{}, {}], "start_dt": "1991-12-05T03:09:58.345Z", "etalon_id": "adipisicing ullamco ad sit commodo", "rel_tp_cd": "Lorem nisi dolore", "access_level": -22902047, "last_update_dt": "1963-07-03T12:01:19.831Z"}, {"citizen": [{}, {}], "start_dt": "1954-04-30T17:57:00.017Z", "etalon_id": "elit dolor voluptate", "rel_tp_cd": "deserunt esse non ut dolore", "access_level": 86139377, "last_update_dt": "1992-11-08T18:44:48.761Z"}, {"citizen": [{}, {}], "start_dt": "1985-07-08T21:58:53.706Z", "etalon_id": "amet ut", "rel_tp_cd": "adipisicing Ut", "access_level": -78533670, "last_update_dt": "1990-08-02T08:24:46.492Z"}, {"citizen": [{}, {}], "start_dt": "1956-06-06T15:05:28.061Z", "etalon_id": "commodo ipsum", "rel_tp_cd": "eu", "access_level": -24463269, "last_update_dt": "2003-05-03T08:19:33.714Z"}], "privilege_control": false, "pname_ext_upd_date": "1982-03-11T07:54:03.689Z", "birth_place_line_one": ""}`)
	srcMap   map[string]interface{}
	srcBytes []byte
	grpcSrc  *structpb.Value
	ji       = jsoniter.ConfigFastest
	bufPool  = sync.Pool{
		New: func() interface{} {
			b := make([]byte, 0, 24*1024)
			return bytes.NewBuffer(b)
		},
	}
)

func init() {
	err := ji.Unmarshal(srcJson, &srcMap)
	if err != nil {
		panic(err)
	}

	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	b := make([]byte, 0, 24*1024)
	buf := bytes.NewBuffer(b)
	err = gob.NewEncoder(buf).Encode(srcMap)
	srcBytes = buf.Bytes()
	if err != nil {
		panic(err)
	}

	grpcSrc = ConvertInterfaceToGrpcStruct(srcMap)
}

func BenchmarkToJsonStd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(srcMap)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkToJsonJsoniter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ji.Marshal(srcMap)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkToBinary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := make([]byte, 0, 24*1024)
		buf := bytes.NewBuffer(b)
		err := gob.NewEncoder(buf).Encode(srcMap)
		if err != nil {
			panic(err)
		}
	}
}

/*func BenchmarkToBinaryWithPool(b *testing.B) {
	bs := make([]byte, 0, 24 * 1024)
	buf := bytes.NewBuffer(bs)
	for i := 0; i < b.N; i++ {
		err := gob.NewEncoder(buf).Encode(srcMap)
		if err != nil {
			panic(err)
		}
	}
}*/

func BenchmarkToGrpcStruct(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ConvertInterfaceToGrpcStruct(srcMap)
	}
}

func BenchmarkFromJsonStd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := make(map[string]interface{})
		err := json.Unmarshal(srcJson, &m)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkFromJsonJsoniter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := make(map[string]interface{})
		err := ji.Unmarshal(srcJson, &m)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkFromBinary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := make(map[string]interface{})
		buf := bytes.NewBuffer(srcBytes)
		err := gob.NewDecoder(buf).Decode(&m)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkFromGrpc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := make(map[string]interface{})
		err := ConvertGrpcToGo(grpcSrc, &m)
		if err != nil {
			panic(err)
		}
	}
}
