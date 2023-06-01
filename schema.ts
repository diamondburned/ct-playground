type blob = null;
type text = null;
type integer = null;
type json<T = unknown> = T;
type proto<T = unknown> = T;
type epoch = null;
type sha256 = null;
type _omitted_ = null;

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

// Some important decisions that we've yet to answer:
//
// 1. Should we keep a blob of the leaf and data together? Or should each small
//    table have its own blob alongside the relevant data?
//
//    This should be decided with future compatibility in mind: if new changes
//    were to be added (e.g., new types) that we don't know about, the blobs
//    should still be stored so we can fix it later without having to re-ingest
//    everything.
//
//    However, if we choose to keep a big blob of (leaf, data), then parsing for
//    smaller information within might be very slow.
//
//    Potential hybrid approach: store a rough protobuf of leaf and data, where
//    anything that isn't a big certificate blob will be parsed into protobuf
//    formats, while unknown/changing data will be raw bytes.
//
// 2. Should we enforce unique constraints on anything, or should we only use
//    indexing?
//
//    It might be unpredictable if a CT log provider decides to troll us and
//    screw up their log entries, but we would also want to enforce some form of
//    indexing for querying performance.
//
// 3. Should CT logs be nested within operators, or does it matter? How often do
//    we need to only know logs from provider X? Are there any advantages to
//    just not nesting?
//
//    Potential nesting: (vendor, serial number, random ID bits)
//
//    Reason for random bits: vendors may screw up and return invalidate
//    certificates with duplicating serial numbers. Adding random bits to that
//    should somewhat help with this. We could also consider using a serially
//    incrementing VARINT if Spanner allows that, since that's more scalable.

export type LogEntry = {
  parentLogID: blob; // references Log.logID
  leaf: {
    type: "timestamped_entry";
    timestampedEntry:
      & {
        timestamp: epoch;
        extensions: json;
        // TODO: jsonData has no discriminator and isn't used for anything in
        // the Go deserialization function. How should we include it?
        jsonData?: json;
      }
      & ({
        logEntryType: "precert";
        precertEntry: {
          issuerKeyHash: sha256;
          tbsCertificate: blob;
        };
      } | {
        logEntryType: "x509";
        x509Entry: null; // TODO
      });
  };
  data: {
    logEntryType: "precert";
    chain: proto<{
      issuerKeyHash: sha256;
      tbsCertificate: proto<X509Certificate>;
      rawASN1: blob;
    }[]>;
  } | {
    logEntryType: "x509";
    chain: proto<{
      certificate: proto<X509Certificate>;
      rawASN1: blob;
    }[]>;
  };
};

export type X509Certificate = {
  version: integer;
  serialNumber: bigint;
  issuer: proto<PKIXName>;
  subject: proto<PKIXName>;
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
  extensions: json;
};

export type PKIXName = _omitted_;
