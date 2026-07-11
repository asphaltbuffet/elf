# SSH public keys as age recipients in config; private identity discovered on-disk

`elf encrypt`/`decrypt` need a key model. age supports its own native keypairs
(`age1…`/`AGE-SECRET-KEY-1…`) as well as reusing existing SSH keys via `filippo.io/age/agessh`. The
native-keypair path would require the user to generate, store, and distribute an age-specific secret
across every machine they decrypt on — a second key-management burden alongside the SSH keys they
already have.

We decided to **reuse SSH keys**. Encryption recipients are a list of **SSH public-key strings**
configured inline in `elf.toml` (a new `recipients` config [[Key]], parsed via
`agessh.ParseRecipient`). Public keys are not secret, so they live in committed config and travel
with the repo: any machine whose public key is listed can decrypt. Decryption discovers the matching
**private** identity on-disk at the conventional path (`~/.ssh/id_ed25519`), overridable with
`--identity`; the private key is **never** placed in config. "Give multiple SSH keys" therefore means
"encrypt to N recipients; any one machine's on-disk private key unlocks it" — accessible from
multiple machines with no per-machine setup beyond the SSH key that is already there.

Recipient parsing is **fail-fast and all-or-nothing**: if the `recipients` list is empty or *any*
entry fails to parse, `encrypt` errors and seals nothing. A partial recipient set is a silent trap —
you would not discover a dropped machine until you were on it with only ciphertext.

## Scope: on-disk keys only for v1

The initial version reads **unencrypted on-disk** private keys only. SSH-agent identities and
passphrase-protected keys (via `agessh.NewEncryptedSSHIdentity`'s passphrase callback) are
**deferred**. A passphrase-protected key therefore fails to parse with a clear error rather than
prompting. This keeps v1 small; the deferred paths are additive and do not change the model above.

## Considered options

- **SSH keys, public in config / private on-disk (chosen).** No new key material to manage;
  recipients are non-secret and committable; multi-machine access is inherent to age's multi-recipient
  encryption.
- **Native age keypairs.** First-class age support, but forces the user to generate and securely
  distribute an age secret to every machine — the exact separate-key-management burden this feature
  set out to avoid.
- **Recipient key *files* by path** (`recipient_files = [...]`, age's `-R`). Rejected: it depends on a
  `.pub` file existing at a fixed path on every machine, reintroducing per-machine setup, whereas
  inline public-key strings are self-contained and travel with the config.

## Consequences

- New public config surface: a `recipients` [[Key]] (list of SSH public-key strings) with a
  `GetRecipients()` getter, and a matching `programs.elf` option in the home-manager module (kept in
  sync with `pkg/config/defaults.go` per the module's contract).
- Decryption depends on an out-of-tree path (`~/.ssh/id_ed25519`), read through the App's `afero.Fs`
  after `~`-expansion so it stays fakeable in tests; `--identity` overrides it.
- Deferring the SSH agent means passphrase-protected keys are unusable until
  `NewEncryptedSSHIdentity` support is added; this is a documented v1 limitation, not a design dead
  end.
