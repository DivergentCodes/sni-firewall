# sni-firewall

A transparent egress filter for TLS and plaintext HTTP. It allows or blocks
each connection by domain, using the SNI of a TLS ClientHello or the `Host`
header of an HTTP request, without decrypting anything.

Domain filtering by SNI trusts a value the client chooses, so a client can send
an allowed name to an attacker's IP. With `--verify`, the filter checks that
the destination actually serves a valid certificate for that name before
relaying the connection.

It started as the defensive half of the
[SNI manipulation PoC](https://github.com/DivergentCodes/sec-poc/tree/main/sni-manipulation),
which demonstrates the bypass in a Terraformed AWS lab.

## How it works

The filter runs on a Linux host that routes traffic for the clients it
protects, such as a NAT instance. `iptables` redirects outbound connections to
its local ports, and it recovers each connection's original destination with
`SO_ORIGINAL_DST`.

For each TLS connection on the TLS port:

1. A connection that does not start with a TLS handshake is closed under
   `--default deny`.
2. The SNI is checked against the domain list. Under `--default deny`, an
   empty or unlisted SNI is blocked with a TLS alert.
3. With `--verify`, the filter connects to the original destination with the
   same SNI and requires a certificate valid for that name. Successful checks
   are cached per name and destination for an hour.
4. The connection is relayed to the original destination unchanged.

With `--http-port`, a second listener applies the same list to the `Host`
header of the first request on each plaintext HTTP connection, answering a
blocked request with `403 Forbidden`. Plaintext has no certificate, so with
`--verify` the original destination must be one of the addresses the `Host`
resolves to from the filter host.

## Usage

```sh
make build
./dist/sni-firewall --domain-file examples/domains.txt --default deny --verify --http-port 3129
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--domain-file` | required | One domain per line. `#` starts a comment. |
| `--default` | required | `deny` treats the list as an allow-list; `allow` treats it as a block-list. |
| `--port` | `3130` | Local port for redirected TLS. |
| `--http-port` | disabled | Local port for redirected plaintext HTTP. |
| `--verify` | off | Verify certificates for TLS, and DNS resolution for HTTP. |

A domain matches exactly. A leading dot, as in `.github.com`, matches any
subdomain but not `github.com` itself.

### Redirecting traffic

On the routing host:

```sh
sysctl -w net.ipv4.ip_forward=1
iptables -t nat -A POSTROUTING -o "$(ip route | awk '/default/ {print $5}')" -j MASQUERADE
iptables -t nat -A PREROUTING -p tcp --dport 443 -j REDIRECT --to-port 3130
iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 3129
```

The filter only sees what is redirected to it. Drop everything else in the
`FORWARD` chain unless you mean to allow it, including UDP 443 so that QUIC
falls back to TCP.

## Limits

- An allowed name hosted where outsiders can obtain a certificate, such as a
  wildcard on a cloud or tunnel provider, passes `--verify`. Keep wildcards
  narrow.
- Domain fronting through an allowed CDN is not detectable without decryption.
- Only IPv4 original destinations are supported.
- Only the first request on a keep-alive HTTP connection is checked.
- The domain list is read once at startup.

## Releases

Pushing a `vX.Y.Z` tag publishes Linux amd64 and arm64 binaries, a
`SHA256SUMS` file, and a build provenance attestation for each binary.

```sh
VERSION=v0.1.0
ARCH=amd64
gh release download "$VERSION" --repo DivergentCodes/sni-firewall \
  --pattern "sni-firewall-${VERSION}-linux-${ARCH}" --pattern SHA256SUMS
sha256sum --check --ignore-missing SHA256SUMS
gh attestation verify "sni-firewall-${VERSION}-linux-${ARCH}" --repo DivergentCodes/sni-firewall
```

## Development

```sh
make lint test build
```

CI runs `golangci-lint`, `go vet`, `go test -race`, Linux builds for amd64 and
arm64, and `govulncheck` on every pull request. `make build` stamps the version
from `git describe`; `--version` prints it.

## License

[AGPL-3.0](LICENSE)
