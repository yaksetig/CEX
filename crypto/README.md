# crypto

The crypto package contains primitives used throughout the project.  It includes
implementations of several timelock puzzle schemes and utilities for producing
cryptographic proofs.

## Proof of Assets

The `provisions` subpackage implements a privacy preserving proof of reserves
protocol.  The exchange commits to the balance of each owned address and then
publishes a Pedersen commitment to the sum of all balances.  Third parties can
challenge the exchange to open individual commitments which proves that the
aggregate commitment matches the on‑chain funds without revealing the complete
set of addresses.
