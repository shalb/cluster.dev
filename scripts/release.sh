#!/bin/bash
set -eux

binDir="bin"
binName="cdev"
eventData=$(cat ${GITHUB_EVENT_PATH})
version=$(echo ${eventData} | jq -r .release.tag_name)
uploadUrl=$(echo ${eventData} | jq -r .release.upload_url)
uploadUrl=${uploadUrl/\{?name,label\}/}

make
cd ${binDir}
for entry in *
do
  echo "${entry}"
  cp "./${entry}/${binName}" ./
  baseName="${binName}-${version}-${entry}"
  arjName="${baseName}.tar.gz"
  checksumName="${baseName}_checksum.txt"
  tar cvfz ${arjName} ${binName}
  rm ${binName}
  
  checksum=$(md5sum ${arjName} | cut -d ' ' -f 1)

  curl \
    -X POST \
    --data-binary @${arjName} \
    -H 'Content-Type: application/octet-stream' \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    "${uploadUrl}?name=${arjName}"

  curl \
    -X POST \
    --data ${checksum} \
    -H 'Content-Type: text/plain' \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    "${uploadUrl}?name=${checksumName}"

  rm -rf "./${arjName}"
done
