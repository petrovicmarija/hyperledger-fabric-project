set -ex # exit on first error

pushd ../test-network
./network.sh down
popd

rm -rf app/wallet/*