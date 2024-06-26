This tool is designed to configure a proxy to return customized responses for specific URLs.

## Features

- **Customizable Responses**: Set responses for specific URLs using one of the following methods:
  - **File**: Return any file stored on the storage.
  - **URL**: Reverse proxy the content of another URL.
  - **Content**: Directly return a string content as the response.
- **Regex Based Matching**:
  - Use regex to route URLs.
  - Dynamically change the reverse proxy destination using variables.
- **Default Route Configuration**: Choose the behavior when no routing matches.
  - Connect to the internet.
  - Connect to another proxy.
  - Deny access.

## Configurations

Configurations are defined using the `config.yaml` file.

| Key             | Type                                   | Description                                                                                                                                                                                                                                             | Example   |
| --------------- | -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- |
| `default_route` | object                                 | Default route configuration.                                                                                                                                                                                                                            | See below |
| `log_level`     | "DEBUG" \| "INFO" \| "WARN" \| "ERROR" | The level of logging detail.                                                                                                                                                                                                                            | `DEBUG`   |
| `always_mitm`   | boolean                                | If `true`, eavesdrop all HTTPS access to get full URL (It may cause performance issues). It may slow down performance<br/> If `false`, only eavesdrop HTTP access if host name is matched (Regex expressions like `.*` in the host name can't be used). | `false`   |
| `routes`        | object                                 | Routing settings.                                                                                                                                                                                                                                       | See below |

### `default_route`

| Key           | Type    | Description                                    | Example                |
| ------------- | ------- | ---------------------------------------------- | ---------------------- |
| `proxy_url`   | string  | The URL of the proxy to connect to by default. | `http://default.proxy` |
| `deny_access` | boolean | Whether to deny access if no routing matches.  | `true`                 |

### `routes`

Define routing settings. Each route is defined in the following format:

| Key        | Type    | Description                                                                             | Example                   |
| ---------- | ------- | --------------------------------------------------------------------------------------- | ------------------------- |
| `url`      | string  | The URL pattern to match. If this URL matches, the specified response will be returned. | `https://example.com/api` |
| `regex`    | boolean | If `true`, regex can be used for URL matching.                                          | `true`                    |
| `response` | object  | The response to return. Choose from `url`, `content`, or `file`.                        | See below                 |

#### `response`

Either a URL, file, or content must be specified.

| Key            | Type   | Description                                           | Example                        |
| -------------- | ------ | ----------------------------------------------------- | ------------------------------ |
| `url`          | object | URL response settings. See below for detailed format. | See below                      |
| `content`      | string | The content to return.                                | `This is the response content` |
| `file`         | string | The file path to return.                              | `/path/to/file`                |
| `status_code`  | int    | The HTTP status code.                                 | `404`                          |
| `content_type` | string | The MIME type of the content.                         | `text/plain`                   |

##### `URL`

| Key         | Type    | Description                                       | Example                      |
| ----------- | ------- | ------------------------------------------------- | ---------------------------- |
| `from`      | string  | The source URL pattern to match.                  | `https://example.com/path`   |
| `to`        | string  | The destination URL to proxy the request to.      | `http://localhost:3000/path` |
| `regex`     | boolean | Whether to use regex for matching the `from` URL. | `true`                       |
| `proxy_url` | string  | The URL of the proxy to use for this route.       | `http://proxy.example.com`   |

## Config example

`config.yaml`

```yaml
default_route:
  deny_access: true

log_level: "INFO"
always_mitm: true

routes:
  # if the request url is "https://example.com/user/[user_id]/post/[post_id]",
  # reverse proxy to "https://example.com/api?user=[user_id]&post=[post_id]".
  - url: "https://example.com/user/[^/]+/post/[^/]"
    regex: true
    response:
      url:
        from: '^https://example\.com/user/([^/]+)/post/([^/]+)'
        to: "https://example.com/api?user=$1&post=$2"
        regex: true
  # if the request url is "https://example.com/not-found",
  # return the content "notfound" with 404 status code.
  - url: "https://example.com/not-found"
    regex: false
    response:
      content: "not found"
      content_type: "text/plain"
      status: 404
  # if the request url is "https://example.com/[any character].png",
  # return the file: "./sample.png"
  - url: 'https://example.com/.*\.png'
    regex: true
    response:
      file: "sample.png"
  # if the request url is "https://example.com/proxy",
  # reverse proxy to "https://example.com/api" using a specific proxy.
  - url: "https://example.com/proxy"
    regex: false
    response:
      url:
        from: "https://example.com/proxy"
        to: "https://example.com/api"
        regex: false
        proxy_url: "http://proxy.example.com"
```
