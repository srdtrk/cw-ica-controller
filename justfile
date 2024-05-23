# Build optimized wasm using the cosmwasm/optimizer:0.15.1 docker image
build-optimize:
  echo "Compiling optimized wasm..."
  docker run --rm -t -v "$(pwd)":/code \
    --mount type=volume,source="$(basename "$(pwd)")_cache",target=/code/target \
    --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
    cosmwasm/optimizer:0.15.1
  echo "Optimized wasm file created at 'artifacts/cw-ica-controller.wasm'"

# Run cargo fmt and clippy
lint:
  cargo fmt --all -- --check
  cargo clippy --all-targets --all-features -- -D warnings

# Build the test contracts using the cosmwasm/optimizer:0.15.1 docker image
build-test-contracts:
  echo "Building test contracts..."
  echo "Building cw-ica-owner..."
  docker run --rm -t -v "$(pwd)":/code \
    --mount type=volume,source="devcontract_cache_burner",target=/code/contracts/burner/target \
    --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
    cosmwasm/optimizer:0.15.1 ./testing/contracts/cw-ica-owner
  echo "Building callback-counter..."
  docker run --rm -t -v "$(pwd)":/code \
    --mount type=volume,source="devcontract_cache_burner",target=/code/contracts/burner/target \
    --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
    cosmwasm/optimizer:0.15.1 ./testing/contracts/callback-counter
  echo "Optimized wasm files created at 'artifacts/cw-ica-owner.wasm' and 'artifacts/callback-counter.wasm'"

# Generate JSON schema files for all contracts in the project
generate-schemas:
  echo "Generating JSON schema files..."
  echo "Generating schema for cw-ica-controller..."
  cargo schema
  echo "Generating schema for cw-ica-owner..."
  cd testing/contracts/cw-ica-owner && cargo schema
  echo "Generating schema for callback-counter..."
  cd testing/contracts/callback-counter && cargo schema

# Run the unit tests
unit-tests:
  RUST_BACKTRACE=1 cargo test --locked --all-features

# Run the e2e tests
e2e-test testname:
  echo "Running {{testname}} test..."
  cd e2e/interchaintestv8 && go test -v -run={{testname}}
