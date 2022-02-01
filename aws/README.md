# Harp AWS

`harp` plugin that allows you to :

* Generate an identity for `sealing` purpose using Envelope encryption with AWS
  KMS;
* Recover AWS KMS protected identity to unseal a container;
* Upload a secret container to S3.

## Build

```sh
export PATH=<harp-repository-path>/tools/bin:$PATH
mage
```

## Sample

### Generate an AWS KMS protected identity

```sh
$ harp aws container identity \
  --description "AWS Recovery" \
  --key-arn "arn:aws:kms:eu-central-1:XXXXXXXXXXXX:key/XXXXXXXXXXXXXXXXXXXXXXX" \
  --out aws-recovery.json
```

If you take a look at the generated file :

```json
{
  "@apiVersion": "harp.elastic.co/v1",
  "@kind": "ContainerIdentity",
  "@timestamp": "2020-11-02T16:41:11.298702Z",
  "@description": "AWS Recovery",
  "public": "WS_fATyjhyeyZld3oEUPmG2trrWoDUdhTVrQvfUeZno",
  "private": {
    "encoding": "kms:aws:RqfOCnUXIIP1u7BaYSANQ_81gY89UMKY-NB1-hvPNrc",
    "content" "ALgBAgIAeOENI.....lzs8Z0Vugsw_6HwNYVFyDY2ZFIGYR"
  }
}
```

It uses the same `ContainerIdentity` format as usual for `passphrase` or
`Vault transit` encryption.

You can seal a container using this identity :

```sh
$ harp container seal
   --in unsealed.bundle \
   --identity $(cat aws.json | jq -r ".public") \
   --out sealed.bundle
```

### Recover a secret container key

```sh
$ harp aws container recover \
  --key-arn "arn:aws:kms:eu-central-1:XXXXXXXXXXXX:key/XXXXXXXXXXXXXXXXXXXXXXX" \
  --identity aws-recovery.json
Container Key: ...
```

### Upload a secret container to S3

> Usual `vault` to `s3` bucket with `harp-server` compatibility workflow with
> `in-transit` encryption (fernet used).

```sh
$ harp keygen fernet > psk.key
$ harp from vault
  --path app/production/security/cloud/v1.0.0
  | harp bundle encrypt --key $(cat psk.key) \
  | harp container seal \
    --identity-file aws-recovery.json \
    --no-container-identity
  | harp aws to s3 \
    --bucket-name harp-containers \
    --object-key sealed.bundle
Container successfully uploaded to: https://harp-containers.s3.eu-central-1.amazonaws.com/sealed.bundle
```

You could specify different endpoint (ex: IBM COS) :

```sh
$ harp aws to s3 \
  --in sealed.bundle \
  --access-key-id $IBM_ACCESS_KEY \
  --secret-access-key $IBM_SECRET_ACCESS_KEY \
  --bucket-name harp-containers-cos-standard-wsa \
  --endpoint s3.eu-de.cloud-object-storage.appdomain.cloud \
  --region eu-de \
  --object-key sealed.bundle
Container successfully uploaded to: https://harp-containers-cos-standard-wsa.s3.eu-de.cloud-object-storage.appdomain.cloud/sealed.bundle
```

Use recovery to export Recovery Container Key :

```sh
$ CONTAINER_KEY=$(harp aws container recover \
  --key-arn "arn:aws:kms:eu-central-1:XXXXXXXXXXXX:key/XXXXXXXXXXXXXXXXXXXXXXX" \
  --identity aws-recovery.json \
  --json | jq -r ".container_key")
```

Expose it using `harp-server` :

```sh
$ export SHUB_SERVER_KEYRING="[$CONTAINER_KEY]"
$ harp server http
  --namespace root:bundle+s3://harp-containers/sealed.bundle
```
