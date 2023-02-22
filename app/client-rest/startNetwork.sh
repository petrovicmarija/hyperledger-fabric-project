set -e # exit on first error

export MSYS_NO_PATHCONV=1
starttime=$(date +%s)
CC_SRC_LANGUAGE=${1:-"go"}
CC_SRC_LANGUAGE=`echo "$CC_SRC_LANGUAGE" | tr [:upper:] [:lower:]`

if [ "$CC_SRC_LANGUAGE" = "go" -o "$CC_SRC_LANGUAGE" = "golang" ] ; then
    CC_SRC_PATH="../chaincode/cars/go"
elif [ "$CC_SRC_LANGUAGE" = "javascript" ] ; then
    CC_SRC_PATH="../chaincode/cars/javascript"
elif [ "$CC_SRC_LANGUAGE" = "java" ] ; then
    CC_SRC_PATH="../chaincode/cars/java"
elif [ "$CC_SRC_LANGUAGE" = "typescript" ] ; then
    CC_SRC_PATH="../chaincode/cars/typescript"
else
    echo The chaincode language ${CC_SRC_LANGUAGE} is not supported.
    echo Supported languages to write chaincode are: go, java, javascript and typescritp.
    exit 1
fi

rm -rf app/wallet/*

pushd ../test-network
./network.sh down
./network.sh up createChannel -ca -s couchdb
./network.sh deployCC -ccn carcc -ccv 1 -cci initLedger -ccl ${CC_SRC_LANGUAGE} -ccp ${CC_SRC_PATH}
popd

cat <<EOF
EOF