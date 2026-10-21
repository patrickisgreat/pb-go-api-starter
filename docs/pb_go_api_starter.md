# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [common_session.proto](#common_session-proto)
    - [Geo](#proto-patrickisgreat-common-session-Geo)
    - [RequestedFeature](#proto-patrickisgreat-common-session-RequestedFeature)
    - [UserSession](#proto-patrickisgreat-common-session-UserSession)
  
- [pb_go_api_starter.proto](#pb_go_api_starter-proto)
    - [GetQuoteFlakyRequest](#proto-patrickisgreat-pb_go_api_starter-api-GetQuoteFlakyRequest)
    - [GetQuoteFlakyResponse](#proto-patrickisgreat-pb_go_api_starter-api-GetQuoteFlakyResponse)
    - [GetQuoteLateRequest](#proto-patrickisgreat-pb_go_api_starter-api-GetQuoteLateRequest)
    - [GetQuoteLateResponse](#proto-patrickisgreat-pb_go_api_starter-api-GetQuoteLateResponse)
    - [GetQuoteRequest](#proto-patrickisgreat-pb_go_api_starter-api-GetQuoteRequest)
    - [GetQuoteResponse](#proto-patrickisgreat-pb_go_api_starter-api-GetQuoteResponse)
  
    - [Quotes](#proto-patrickisgreat-pb_go_api_starter-api-Quotes)
  
- [Scalar Value Types](#scalar-value-types)



<a name="common_session-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## common_session.proto



<a name="proto-patrickisgreat-common-session-Geo"></a>

### Geo
the geographic location, determined from IP geolocation.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| country_code | [string](#string) |  | two-letter country code, or &#39;--&#39; if unknown |
| city | [google.protobuf.StringValue](#google-protobuf-StringValue) |  |  |
| region | [google.protobuf.StringValue](#google-protobuf-StringValue) |  |  |






<a name="proto-patrickisgreat-common-session-RequestedFeature"></a>

### RequestedFeature



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="proto-patrickisgreat-common-session-UserSession"></a>

### UserSession
Metadata about the interaction between a user and the system


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_urn | [google.protobuf.StringValue](#google-protobuf-StringValue) |  | user making the request |
| agent_urn | [google.protobuf.StringValue](#google-protobuf-StringValue) |  | the property being used for this session |
| geo | [Geo](#proto-patrickisgreat-common-session-Geo) |  | the geographic location of the user |
| scopes | [string](#string) | repeated | OAuth scopes for this session |
| features | [string](#string) | repeated | business features that are enabled for this session |
| app_variant_ids | [int32](#int32) | repeated | propagate variants for A/B testing |
| app_requested_features | [RequestedFeature](#proto-patrickisgreat-common-session-RequestedFeature) | repeated | features requested in A/B tests |





 

 

 

 



<a name="pb_go_api_starter-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## pb_go_api_starter.proto



<a name="proto-patrickisgreat-pb_go_api_starter-api-GetQuoteFlakyRequest"></a>

### GetQuoteFlakyRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_session | [proto.patrickisgreat.common.session.UserSession](#proto-patrickisgreat-common-session-UserSession) |  | The session of the user requesting a quote. |
| fail | [bool](#bool) |  | Setting to true causes the request to fail for sure (i.e., not to just randomly fail) |






<a name="proto-patrickisgreat-pb_go_api_starter-api-GetQuoteFlakyResponse"></a>

### GetQuoteFlakyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| quote | [string](#string) |  | A random quote from the embedded quote collection |






<a name="proto-patrickisgreat-pb_go_api_starter-api-GetQuoteLateRequest"></a>

### GetQuoteLateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_session | [proto.patrickisgreat.common.session.UserSession](#proto-patrickisgreat-common-session-UserSession) |  | The session of the user requesting a quote. |
| delay_ms | [google.protobuf.Int32Value](#google-protobuf-Int32Value) |  | Number of milliseconds to wait before returning the response |






<a name="proto-patrickisgreat-pb_go_api_starter-api-GetQuoteLateResponse"></a>

### GetQuoteLateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| quote | [string](#string) |  | A random quote from the embedded quote collection |






<a name="proto-patrickisgreat-pb_go_api_starter-api-GetQuoteRequest"></a>

### GetQuoteRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_session | [proto.patrickisgreat.common.session.UserSession](#proto-patrickisgreat-common-session-UserSession) |  | The session of the user requesting a quote. |






<a name="proto-patrickisgreat-pb_go_api_starter-api-GetQuoteResponse"></a>

### GetQuoteResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| quote | [string](#string) |  | A random quote from the embedded quote collection

Note: a quote is always returned, so no string wrapper type is needed |





 

 

 


<a name="proto-patrickisgreat-pb_go_api_starter-api-Quotes"></a>

### Quotes


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetQuote | [GetQuoteRequest](#proto-patrickisgreat-pb_go_api_starter-api-GetQuoteRequest) | [GetQuoteResponse](#proto-patrickisgreat-pb_go_api_starter-api-GetQuoteResponse) | Get a random quote. |
| GetQuoteFlaky | [GetQuoteFlakyRequest](#proto-patrickisgreat-pb_go_api_starter-api-GetQuoteFlakyRequest) | [GetQuoteFlakyResponse](#proto-patrickisgreat-pb_go_api_starter-api-GetQuoteFlakyResponse) | Get a random quote. Fails roughly half the time, or always when `fail` is set. Useful for exercising error handling and alerting. |
| GetQuoteLate | [GetQuoteLateRequest](#proto-patrickisgreat-pb_go_api_starter-api-GetQuoteLateRequest) | [GetQuoteLateResponse](#proto-patrickisgreat-pb_go_api_starter-api-GetQuoteLateResponse) | Get a random quote after the specified delay (random 100-1000ms when unset). Useful for exercising latency alerting. |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

