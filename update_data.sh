#!/bin/sh

set -e 
bq query --use_legacy_sql=false --format=json --max_rows=100000 < query_intersections.sql > temp.json

function rewrite() {
    local county=$1
    local output_key=$2
    local target_file="${output_key}_intersecting_streets.json"
    local target_dir="../www/parking_enforcement"
    echo "extracting ${county} to ${target_file}"
    jq -c "[.[] | select(.county==\"${county}\") | {c:.county, st:.street, i:.intersections, a:.agencies, n:.number_violations}]" temp.json > ${target_dir}/${target_file}
    echo "${county}: extracted $(jq -c '. | length' ${target_dir}/${target_file}) records"
    return $?
}

echo "Got $(jq -c '. | length' temp.json) records"

rewrite "Manhattan" "ny"
rewrite "Bronx" "bx"
rewrite "Brooklyn" "bk"
rewrite "Queens" "qn"
rewrite "Staten Island" "si"

