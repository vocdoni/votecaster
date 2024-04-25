#¡/bin/bash
REPOSITORY=${REPOSITORY:-https://github.com/vocdoni/degen-communities.git}
CONTRACT=${CONTRACT:-src/CommunityHub.sol}
PKG=${PKG:-CommunityHubToken}
OUTPUT=${OUTPUT:-communityhubtoken/communityhubtoken.go}

# clone the repository
echo "Cloning repository ${REPOSITORY}"
git clone --recursive $REPOSITORY temp
cd temp
# replace openzeppelin-contracts with relative path to lib/openzeppelin-contracts
contract_dir=$(dirname "${CONTRACT}")
echo "Replacing openzeppelin-contracts with relative path in ${contract_dir} files"
find "$contract_dir" -type f -exec sed -i '' 's|openzeppelin-contracts/\(.*\)|../lib/openzeppelin-contracts/\1|g' {} +
# compile the contract to get the abi
echo "Compiling contract ${CONTRACT}"
solc --abi --bin ${CONTRACT} -o build
cd ..
# get the abi basename
filename="${CONTRACT##*/}"
basename="${filename%.*}"
# generate the go binding
output_dir=$(dirname "${OUTPUT}")
mkdir -p $output_dir
echo "Generating go binding for ${basename} in ${output_dir}"
abigen --abi temp/build/${basename}.abi --pkg ${PKG} --out ${OUTPUT}
# cleanup
rm -rf temp