#!/bin/sh

if ! [ -e temp.json ]; then
    bq query --use_legacy_sql=false --format=json --max_rows=200000 < query_intersections.sql > temp.json
    if [ $? != 0 ]; then
        echo "error querying bigquery"
        exit 1
    fi
else
    echo "using existing temp.json file"
fi

FY="fy_2023"

function rewrite() {
    local county=$1
    local output_key=$2
    local target_file="${FY}_${output_key}_intersecting_streets.tsv"
    local target_dir="www"
    echo "extracting ${county} to ${target_file}"
    jq -c ".[] | select(.county==\"${county}\") | {street, intersections:(.intersections | join(\"|\")), agencies:(.agencies|join(\"|\")), number_violations}" temp.json | json2csv -d "$(printf "\t")" -p -k street,intersections,agencies,number_violations > ${target_dir}/${target_file}
    echo "${county}: extracted $(cat ${target_dir}/${target_file} | wc -l ) records"
    return $?
}
set -e 

echo "Got $(jq -c '. | length' temp.json) records"

rewrite "Manhattan" "ny"
rewrite "Bronx" "bx"
rewrite "Brooklyn" "bk"
rewrite "Queens" "qn"
rewrite "Staten Island" "si"

