## Package Architecture

The main request path runs from the proxy through route matching and middleware to one of the response handlers.

```mermaid
flowchart LR
    Main["main<br/>CLI and HTTP server"]
    Proxy["proxy<br/>HTTP(S), MITM, request lifecycle"]
    Routers["routers<br/>Route matching"]
    Middlewares["middlewares<br/>Transform, headers, status"]

    subgraph Responders["Response handlers (routers)"]
        Content["ContentResponder"]
        File["FileResponder"]
        ReverseProxy["ReverseProxyTransport"]
    end

    Main --> Proxy
    Proxy --> Routers
    Routers --> Middlewares
    Middlewares --> Content
    Middlewares --> File
    Middlewares --> ReverseProxy

    Cache["utils/cache<br/>TLS certificate cache"]
    Proxy -.-> Cache
```

When making changes:

- Update `models` and `models/config-spec.json` for configuration schema changes.
- Update `routers` for route matching or Content, File, and Rewrite behavior.
- Update `middlewares` for response transformations and common response overrides.
- Update `proxy` for HTTP proxying, HTTPS interception, certificates, and request lifecycle behavior.
- Add end-to-end coverage under `integration_test`.
