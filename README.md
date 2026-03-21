# net-tools

net-tools is a collection of small, single-purpose networking utilities. Each tool is a simple wrapper around Go's standard library functions (and external libraries), providing easy-to-use command-line applications for common networking tasks.

## Available Tools.

- **http-server**: Serve files from a directory over HTTP.
- **dns-resolver**: Resolve multiple DNS queries from the command line at once.
- **mac-changer**: Change the MAC address of a network interface.
- **mail-client**: Send emails from the command line.
- **arpspoofer**: Perform ARP spoofing on a local network.
- **Extensible**: Easily add more networking utilities in the future.

## Installation

Clone the repository and build the tools:

```sh
git clone https://github.com/kakeetopius/net-tools.git
cd net-tools
```
Build a particular tool:

```sh
go build ./cmd/tool_name
```
For Example
```sh
go build ./cmd/http-server
```

## Usage

Each tool is a standalone command-line application. For example, to start the HTTP server:

```sh
./http-server -dir /path/to/serve -port 8080
```

Refer to each tool's `-h` or `--help` flag for usage instructions:

```sh
./arpspoofer --help
```

## Requirements

- Go 1.18 or newer
- Linux (tested), should work on other Unix-like systems

## Contributing

Contributions are welcome! Feel free to submit pull requests for new utilities or improvements to existing ones.

## License
MIT License. See [LICENSE](LICENSE) for details.
