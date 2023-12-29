SELECT
CASE
  WHEN violation_county in ("MN", "NY") THEN "Manhattan"
  WHEN violation_county in ("BX", "Bronx") THEN "Bronx"
  WHEN violation_county in ("BK", "K", "Kings") THEN "Brooklyn"
  WHEN violation_county in ("Q", "QN", "Qns") THEN "Queens"
  WHEN violation_county in ("R", "Rich", "ST") THEN "Staten Island"
  ELSE "Unknown"
END as county,
  street_name as street,
  ARRAY_AGG(distinct regexp_replace(LOWER(intersecting_street), "((^[^/]+$)|(.+ (.+/of .*)))", "\\2\\4") IGNORE NULLS ) as intersections,
  ARRAY_AGG(distinct issuing_agency) as agencies,
  count(*) as number_violations
FROM `jehiah-p1.opendata.parking_violations_fy`
WHERE 
issuing_agency in ('T','P','S')
and issue_date is not null
and issue_date between '2022-10-01' and CURRENT_DATE
and street_name is not null
and violation_county is not null
and violation_county in ("MN", "NY", "BX", "Bronx", "BK", "K", "Kings", "Q", "QN", "Qns", "R", "Rich", "ST")
group by 1, 2
having number_violations > 3
order by 1, lower(street_name)
