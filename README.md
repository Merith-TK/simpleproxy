# SimpleProxy Usage Guide

## Overview

`SimpleProxy` is a lightweight proxy tool that uses a configuration file to define proxy settings. The configuration file, `config.json`, is **required** for the proxy to function.

## Usage

To run SimpleProxy, execute the following command in your terminal:

```sh
simpleproxy config.json
```

### Configuration File

The `config.json` file is essential for SimpleProxy to know which ports to bind to and which remote addresses to proxy from. Below is an example of the configuration file:

```json
{
    "proxy": [
        /*
        {
            "local" : ":port",      // Port to bind to, defaults to remote port
            "remote": "addr:port",  // Remote address to proxy from
            "type"  : "both"        // Protocol type: tcp, udp, or both; defaults to both
        }
        */
        {
            "remote": "example.com:25565"
        },
        {
            "remote": "example.com:42420"
        }
    ]
}
```

### Configuration Options

Each entry in the `proxy` array represents a proxy rule with the following options:

- **local**: Optional. The port to bind to locally. If omitted, the proxy will use the port specified in the `remote` field.
- **remote**: Required. The remote address and port to proxy from.
- **type**: Optional. Specifies the protocol type (`tcp`, `udp`, or `both`). Defaults to `both` if not specified.

### Example Configuration

Here are a couple of examples based on the configuration options:

1. **Basic Configuration:**

   ```json
   {
       "proxy": [
           {
               "remote": "example.com:25565"
           },
           {
               "remote": "example.com:42420"
           }
       ]
   }
   ```

   In this configuration, SimpleProxy will bind to ports `25565` and `42420` locally, and proxy traffic to `example.com` on the same ports.

2. **Custom Local Port and Protocol Type:**

   ```json
   {
       "proxy": [
           {
               "local": ":8080",
               "remote": "example.com:25565",
               "type": "tcp"
           },
           {
               "remote": "example.com:42420",
               "type": "udp"
           }
       ]
   }
   ```

   In this configuration:
   - The first proxy rule binds to port `8080` locally and proxies TCP traffic to `example.com:25565`.
   - The second proxy rule binds to port `42420` locally (same as the remote port) and proxies UDP traffic to `example.com:42420`.

## Running the Proxy

After setting up the `config.json` file with your desired proxy rules, run SimpleProxy using the command mentioned earlier:

```sh
simpleproxy config.json
```

This will start the proxy based on the defined configuration. Make sure the `config.json` file is in the same directory as the `simpleproxy` executable or provide the correct path to the configuration file.
