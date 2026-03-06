function parseDate(dateString) {
    return new Date(dateString + " UTC");
}

export const FYLookup = {
        "FY2026": {
        "dataset": "pvqr-7yc4",
        "start": parseDate("2025-07-01"),
        "end": {
            "P":parseDate("2026-01-11"),
            "T":parseDate("2026-01-27"),
            "S":parseDate("2026-01-29")
        }
        },
        "FY2025": {
        "dataset":"m5vz-tzqv", 
        "start": parseDate("2024-07-01"),
        "end": parseDate("2025-07-01"),
        },
        "FY2024": {
        "dataset":"8zf9-spf8", 
        "start": parseDate("2023-07-01"),
        "end": parseDate("2024-07-01"),
        },
        "FY2023": {
        "dataset":"869v-vr48", 
        "start": parseDate("2022-07-01"),
        "end": parseDate("2023-06-30")
        },
        "FY2022": {
        "dataset":"pvqr-7yc4",
        "start": parseDate("2022-07-01"),
        "end": parseDate("2023-07-01")
        },
        "FY2021": {
        "dataset":"7mxj-7a6y",
        "start": parseDate("2021-07-01"),
        "end": parseDate("2022-06-30")
        }
    }
    