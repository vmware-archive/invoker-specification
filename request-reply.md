# Request / Reply Interaction Model

## Lifecycle

## Invocation

## Request

## Reply

[//]: # (Comment: The following section also appears in request-reply.md)

## Supported MIME Types
An invoker SHOULD support the following MIME types, both when dealing with receiving data and when asked to serialize data back to the streaming processor:
* `text/plain`: when receiving data tagged with this content type and a function argument expects a "string", then an invoker MUST be capable of fulfilling that value. Conversely, when asked to produce that content type and receiving a "string" from the function, an invoker MUST be able to serialize the string using that MIME type.
Additionally, when dealing with "byte arrays" in the function signature, an invoker SHOULD be able to serialize/deserialize from/to a value using this MIME type, honoring the value of the `encoding` MIME type parameter if present.
* `application/json`: when receiving data tagged with this content type, an invoker SHOULD attempt to map the JSON structure to the function argument using idiomatic behavior from the target runtime. This MAY involve using general purpose data structures (*e.g.* maps or dictionaries), or trying to map content to structured data. Conversely, when asked to produce that content type, an invoker SHOULD use idiomatic conventions of the target runtime to serialize JSON. The behavior when encountering missing or extra fields, or relative to circular references is beyong the scope of this document and is left at the discretion of the invoker.
* `application/octet-stream`: when receiving data tagged with this content type and a function argument expects a "byte array", then an invoker MUST be capable of fulfilling that value, passing the `payload` as-is. Conversely, when asked to produce that content type and receiving a "byte array" from the function, an invoker MUST be able to produce a payload.

In addition, an invoker SHOULD provide a way for the function to extend the set of supported MIME types, *e.g.* by providing an extension mechanism to register additional "handlers". The specific details of such a mechanism are beyond the scope of this document.

## Error Conditions

## Automatic Support for Request/Reply in the Streaming Case
