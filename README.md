# Flexy Proxy

An easy-to-start, YAML-based flexible proxy for software development. Return customized responses for specific URLs.

## Features

- **Customizable Responses**: Set responses for specific URLs using one of the following methods:
  - **File**: Return any file stored on the storage.
  - **Rewrite**: Rewrite the URL to another URL like a reverse proxy.
  - **Content**: Directly return a string content as the response.
  - **Transform**: Apply a transformation command to the response content.
- **Regex Based Matching**:
  - Use regex to route URLs.
  - Dynamically change the reverse proxy destination using variables.
- **Default Route Configuration**: Choose the behavior when no routing matches.
  - Connect to the internet.
  - Connect to another proxy.
  - Deny access.

## Installation

### Binary Download

You can download the binary files from the [GitHub Releases](https://github.com/kajikentaro/flexy-proxy/releases) page. Choose the version that suits your operating system and architecture.

### Build from Source with Go

If you prefer to build from source, you can install Flexy Proxy using `go install`.

```
go install github.com/kajikentaro/flexy-proxy@latest
```

## Usage

1. Prepare the executable and add it to your PATH  
   Make sure the `flexy` binary is either in your system's PATH or moved to a location like `/usr/local/bin`.

2. Create a `config.yaml` file  
   Define your routes and responses in a `config.yaml` file. For example:

   ```yaml
   routes:
     - url: "http://sample.test"
       regex: false
       response:
         content: "hello world\n"
   ```

3. Run the proxy with your config file  
   Execute `flexy` with the `-f` flag pointing to your configuration file:

   ```bash
   flexy -f config.yaml
   ```

4. Use the proxy  
   Now, you can use the proxy with a tool like `curl`:

   ```bash
   $ curl http://sample.test -x http://localhost:8888
   hello world
   ```

## Configurations

| Key               | Type                                   | Description                                                                                                                                                                                                              | Example      |
| ----------------- | -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------ |
| `default_route`   | object                                 | Default route configuration.                                                                                                                                                                                             | See below    |
| `log_level`       | "DEBUG" \| "INFO" \| "WARN" \| "ERROR" | The level of logging detail.                                                                                                                                                                                             | `DEBUG`      |
| `always_mitm`     | boolean                                | If `true`, eavesdrop all HTTPS access to get full URL. (It may slow down performance) <br/> If `false`, only eavesdrop HTTP access if host name is matched. (Regex expressions like `.*` in the host name can't be used) | `false`      |
| `routes`          | object                                 | Routing settings.                                                                                                                                                                                                        | See below    |
| `certificate`     | string                                 | The path to the SSL certificate file.                                                                                                                                                                                    | `server.crt` |
| `certificate_key` | string                                 | The path to the private key file corresponding to `certificate`.                                                                                                                                                         | `server.key` |

### `default_route`

| Key           | Type    | Description                                    | Example                |
| ------------- | ------- | ---------------------------------------------- | ---------------------- |
| `proxy`       | string  | The URL of the proxy to connect to by default. | `http://default.proxy` |
| `deny_access` | boolean | Whether to deny access if no routing matches.  | `true`                 |

### `routes`

Define routing settings. Each route is defined in the following format:

| Key        | Type    | Description                                                                                 | Example                   |
| ---------- | ------- | ------------------------------------------------------------------------------------------- | ------------------------- |
| `url`      | string  | The URL pattern to match. If this URL matches, the specified response will be returned.     | `https://example.com/api` |
| `regex`    | boolean | If `true`, regex can be used for URL matching.                                              | `true`                    |
| `response` | object  | The response to return. `rewrite`, `content`, or `file` and other options can be specified. | See below                 |

#### `response`

Only one of `rewrite`, `file`, or `content` can be specified.

| Key            | Type   | Description                                        | Example                              |
| -------------- | ------ | -------------------------------------------------- | ------------------------------------ |
| `rewrite`      | object | Rewrite settings. See below for detailed format.   | See below                            |
| `content`      | string | The content to return.                             | `This is the response content`       |
| `file`         | string | The file path to return.                           | `/path/to/file`                      |
| `status`       | int    | The HTTP status code.                              | `404`                                |
| `content_type` | string | The MIME type of the content.                      | `text/plain`                         |
| `headers`      | map    | The additional headers to include in the response. | `"Access-Control-Allow-Origin": "*"` |
| `transform`    | string | The command to transform the response content.     | `sed -E 's/foo/bar/g'`               |

##### `rewrite`

| Key     | Type    | Description                                                                                                                                         | Example                      |
| ------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------- |
| `from`  | string  | The pattern to be replaced.                                                                                                                         | `https://example.com/path`   |
| `to`    | string  | The replacement string.                                                                                                                             | `http://localhost:3000/path` |
| `regex` | boolean | Whether to use regex for matching the `from` URL.                                                                                                   | `true`                       |
| `proxy` | string  | The URL of the proxy to use for this route. If not specified, `default_route.proxy` will be used. To disable the proxy, specify an empty string "". | `http://proxy.example.com`   |

## Config example

`config.yaml`

```yaml
default_route:
  deny_access: true

log_level: "INFO"
always_mitm: true

routes:
  # if the request URL is "https://example.com/user/[user_id]/post/[post_id]",
  # reverse proxy to "https://example.com/api?user=[user_id]&post=[post_id]".
  - url: "https://example.com/user/[^/]+/post/[^/]"
    regex: true
    response:
      rewrite:
        from: '^https://example\\.com/user/([^/]+)/post/([^/]+)'
        to: "https://example.com/api?user=$1&post=$2"
        regex: true
  # if the request URL is "https://example.com/not-found",
  # return the content "not found" with 404 status code.
  - url: "https://example.com/not-found"
    regex: false
    response:
      content: "not found"
      content_type: "text/plain"
      status: 404
  # if the request URL is "https://example.com/[any character].png",
  # return the file: "./sample.png"
  - url: 'https://example.com/.*\.png'
    regex: true
    response:
      file: "sample.png"
  # if the request URL is "https://example.com/proxy",
  # reverse proxy to "https://example.com/api" using a specific proxy.
  - url: "https://example.com/proxy"
    regex: false
    response:
      rewrite:
        from: "https://example.com/proxy"
        to: "https://example.com/api"
        regex: false
        proxy: "http://proxy.example.com"
  # if the request URL is "https://content.test",
  # return the content "basic" with a custom header.
  - url: "https://content.test"
    regex: false
    response:
      content: "basic"
      headers:
        "Access-Control-Allow-Origin": "*"
  # if the request URL is "https://content.test/",
  # return the content "foo" transformed to "bar".
  - url: "https://content.test/"
    regex: false
    response:
      content: "foo"
      transform: "sed -E 's/foo/bar/g'"
```

## Certificates

The following commands create a private key (`server.key`) and a certificate (`server.csr`). By specifying options, Flexy Proxy can use this private key and certificate to generate new certificates for the requested hostname and use them for communication. By installing the `server.csr` certificate on your PC or browser, responses from Flexy Proxy will be considered secure.

```
openssl genrsa -out server.key
openssl req -new -key server.key -out server.csr -subj "/C=JP/ST=Tokyo/L=Minato/O=Example Company/OU=IT Department/CN=example.com"
openssl x509 -req -days 3650 -signkey server.key -in server.csr -out server.crt
```

Specify the certificates in the YAML file as follows:

```yaml
certificate: "server.crt"
certificate_key: "server.key"
```
