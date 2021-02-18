# Yubikey managed container identity

## Prepare your yubikey

Generate a `secp256r1` EC private key in the yubikey :

```sh
yubico-piv-tool --slot=82 --algorithm=ECCP256 --touch-policy=always --pin-policy=once -a generate -o recovery.pub.pem
```

Generate a `self-signed` certificate with the required `Organization` attribute set to `harp-plugin-yubikey`

```sh
yubico-piv-tool --slot=82 -a verify-pin -a selfsign-certificate --subject='/CN=Recovery Harp/O=harp-plugin-yubikey/' --valid-days=3650 -i recovery.pub.pem -o recovery.cert.pem
```

Import the certificate in the yubikey :

```sh
yubico-piv-tool --slot=82 -a import-certificate -i recovery.cert.pem
```

## Create your recovery identity

```sh
$ harp-yubikey container identity --serial $YK_SERIAL --slot 82 --description="Recovery from Yubikey"  | jq
{
  "@apiVersion": "harp.elastic.co/v1",
  "@kind": "ContainerIdentity",
  "@timestamp": "2021-02-15T18:20:58.929526Z",
  "@description": "Recovery from Yubikey",
  "public": "J8QzwQwUIrS2VQtNbzp5bCT5jhHBn6aXXQ2-CWhsigc",
  "private": {
    "encoding": "piv:yubikey:000000000:82:DkDc7g",
    "content": "... REDACTED ..."
  }
}
```

## Recover container key

```sh
$ harp-yubikey container recover --identity id.json
Enter PIN for Yubikey with serial 000000000:
# Don't forget to touch the key (according to defined private key TouchPolicy)
Container key : luCo-1RSFdvXUVLLNyiytc8vEZFutBK1XG_NsuAVT-4
```
