# Request / Reply Interaction Model
When running in request / reply mode, a function invoker MUST behave as an http server and invoke its function in a synchronous manner.

That server can be invoked by any http client that adheres to the specification below. Because basic http is inherently request / reply oriented, this interaction model MUST only be used to drive functions that accept exactly one parameter and produce exactly one result: functions that either accept several parameters or produce several results only make sense in the context of asynchronous [streaming](streaming.md), because if parameters (respectively results) were accepted (respectively produced) in a synchronous manner, then they could be regrouped in a compound parameter (respectively result).



## Lifecycle
When a request / reply invoker starts, it MUST start an http/2 server listening on the port defined by the `PORT` environment variable (it MUST fall back on port 8080 if that variable is not defined) and accept POST requests on `/`. It MAY load the function early although an invoker MAY also assume that user functions may be poorly written and may maintain (mutable) state when they shouldn't, and hence MAY decide to re-load the function at each invocation.

## Invocation
Invocations of the invoker by an http client MAY happen concurrently and the invoker MUST maintain isolation of state inside the function as much as possible given the target runtime. Each http request MUST trigger exactly one invocation of the function (minus error cases).

Function authors MAY decide to maintain state at the function level (as opposed to *per-invocation*) but it is not the responsibility of the invoker to protect against race conditions in those cases.

## Request
To successfully trigger an invocation, an http request MUST adhere to the following constraints. An invoker MAY honor additional behavior (in particular, react to additional headers or header values) but this is not required.

An invoker MUST only consider POST requests on `/` for invocations. Other methods or other paths MUST NOT lead to a function invocation. In those cases, an invoker MAY respond with an error condition or an unrelated successful payload.

An incoming http request SHOULD contain a `Content-Type` header that will be used by the invoker to drive deserialization of the http request *body*. If the http request does not bear such a header, an invoker MUST interpret the body as having an `application/octet-stream` content-type.

An incoming http request SHOULD contain an `Accept` header that will be used by the invoker to drive serialization of the invocation result back to the client. In the absence of such a header, an invoker MUST interpret the client as accepting an `*/*` response.

Upon reception of a well formed request, an invoker MUST deserialize the request body according to its `Content-Type` and any hint it can gather from the function. If possible, an invoker SHOULD consider the type accepted by the function signature to drive the deserialization process. If an invoker can't decide how to deserialize the request body, it SHOULD reply with an http 415 error code. If an error happens while deserializing, an invoker MUST reply with a 5xx error code.

At this point, an invoker MUST attempt to invoke the function with the result of deserialization.

## Reply
Upon successful invocation of the function, an invoker MUST attempt to serialize the result as an http response. What "successful" means depends on the target runtime and the function and its exact specification is beyond the scope of this specification. An invoker SHOULD interpret success / failure in an idiomatic way to the best extent possible though.

In case of success, an invoker MUST select a MIME type according to the semantics of the `Accept` header, as defined in [rfc 2616](https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html) and serialize the function result using that MIME type, and it MUST set the reponse's `Content-Type` header to the selected MIME type. A successful invocation MUST result in a 200 response code. If no MIME type can be selected, an invoker SHOULD respond with an http 406 error code. If an error happens while serializing the result, it must reply with a 5xx error code.

In case of unsuccessful function invocation, an invoker MUST reply with an http 5xx error code and it MAY provide details (such as a stacktrace) in the body of the http response.


[//]: # (Comment: The following section also appears in streaming.md)

## Supported MIME Types
An invoker SHOULD support the following MIME types, both when dealing with receiving data and when asked to serialize data back to the streaming processor:
* `text/plain`: when receiving data tagged with this content type and a function argument expects a "string", then an invoker MUST be capable of fulfilling that value. Conversely, when asked to produce that content type and receiving a "string" from the function, an invoker MUST be able to serialize the string using that MIME type.
Additionally, when dealing with "byte arrays" in the function signature, an invoker SHOULD be able to serialize/deserialize from/to a value using this MIME type, honoring the value of the `encoding` MIME type parameter if present.
* `application/json`: when receiving data tagged with this content type, an invoker SHOULD attempt to map the JSON structure to the function argument using idiomatic behavior from the target runtime. This MAY involve using general purpose data structures (*e.g.* maps or dictionaries), or trying to map content to structured data. Conversely, when asked to produce that content type, an invoker SHOULD use idiomatic conventions of the target runtime to serialize JSON. The behavior when encountering missing or extra fields, or relative to circular references is beyong the scope of this document and is left at the discretion of the invoker.
* `application/octet-stream`: when receiving data tagged with this content type and a function argument expects a "byte array", then an invoker MUST be capable of fulfilling that value, passing the `payload` as-is. Conversely, when asked to produce that content type and receiving a "byte array" from the function, an invoker MUST be able to produce a payload.

In addition, an invoker SHOULD provide a way for the function to extend the set of supported MIME types, *e.g.* by providing an extension mechanism to register additional "handlers". The specific details of such a mechanism are beyond the scope of this document.

## <a name="support-for-streaming-functions"></a>Support for Streaming Functions
Some subset of streaming functions may be made invocable using the request / reply invocation model. One *way* to achieve that using buildpacks is detailed in the [packaging](packaging.md) document, but the *semantics* of the conversion are explained here and MUST be honored by any invoker that claims to support invocation of streaming functions using the request / reply interaction model.

To be invocable using this interaction model, a streaming function MUST accept exactly one streaming input and produce exactly one streaming output. Upon reception of an incoming http request satisfying the prerequisites exposed above, a streaming rpc invocation MUST be made with the following attributes:
* A `StartFrame` MUST be sent with a list of `expectedContentTypes` of size 1, whose value is the value of the `Accept` header of the http request.
* An `InputFrame` MUST be sent with its `payload` set as the body of the http request, its `contentType` field set as the value of the `Content-Type` http header. Other http headers MAY be forwarded using the `headers` field of the `InputFrame`. The `Content-Type` and `Accept` http headers MUST NOT be forwarded this way.
* The input stream of the invocation must then be [completed](glossary.md#stream-completion)
* Upon reception of a single `OutputFrame` followed by the completion of the output stream, the contents of the `OutputFrame` MUST be forwarded as an http response as such: its `payload` MUST form the response body, its `contentType` MUST be set as the http `Content-Type` header and some or all of its additional `headers` MAY be copied to http headers. Although it should not be present, a custom `header` whose name could clash with the http `Content-Type` header MUST NOT be forwarded back.
* Any reception of an `OutputFrame` before the input stream has been completed, or any reception of an additional `OutputFrame` after the first (*i.e.* absence of completion signal on the output stream) MUST result in an http 5xx error condition.
* Any reception of an error signal in the output stream MUST result in an http 5xx error condition.
