type blob = null;
type text = null;
type integer = null;
type json<T = unknown> = T;
type epoch = null;
type sha256 = null;

export type Log = {
  logID: blob; // primary key
  pubKey: blob;
  apiURL: text;
  mmd: integer;
  status: integer;
  type: "prod" | "test";
  state: json;
  previousOperators: json;
};

export type LogEntry = {
  parentLogID: blob; // references Log.logID
  leaf: {
    type: "timestamped_entry";
    timestampedEntry:
      & {
        timestamp: epoch;
        jsonData: json;
        extensions: json;
      }
      & ({
        entryType: "precert";
        precertEntry: {
          issuerKeyHash: sha256;
          tbsCertificate: blob;
        };
      } | {
        entryType: "x509";
        x509Entry: null; // TODO
      });
  };
  x509cert: null; // TODO
  precert: {
    // submitted contains DER bytes of an ASN.1 Certificate
    submitted: { asn1_der: blob };
    issuerKeyHash: sha256;
    tbsCertificate: blob;
  };
};

export type X509Certificate = {
  version: integer;
  serialNumber: bigint;
  issuer: PKIXName;
  subject: PKIXName;
  notBefore: epoch;
  notAfter: epoch;
  keyUsage: integer; // enum of some sort
  signature: {
    algorithm: "md2-rsa" | "md5-rsa" | "sha1-rsa" | "sha256-rsa"; // ...
    data: blob;
  };
  publicKey: {
    algorithm: "rsa" | "dsa" | "ecdsa" | "ed25519" | "rsaesoaep";
    data: blob;
  };
  extensions: any;
};

// Taken from x509/pkix.Name.
export type PKIXName = {
  country: text[];
  organization: text[];
  organizationalUnit: text[];
  locality: text[];
  province: text[];
  streetAddress: text[];
  postalCode: text[];
  serialNumber: text;
  commonName: text;
  names: ASN1AttributeTypeAndValue[];
};

export type ASN1ObjectIdentifier = integer[]; // encode as 0.1.2 is ok too

export type ASN1AttributeTypeAndValue = {
  type: ASN1ObjectIdentifier;
  value: json;
};
