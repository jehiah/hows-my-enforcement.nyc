#!/bin/sh

curl -s https://www.nyc.gov/assets/nypd/data/precinct-house.json | \
jq '[ .features[] | .properties | {Precinct:.PRECINCT, Borough:.BORO, Desc: "\(.BORO) - \(.NUM) \(.STREET)"} ]' > precincts.json