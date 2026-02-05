# Endpoint Version Analysis

⚠️ Found 2 endpoint version mismatches:

## healthchecks

| Source | Endpoint | Version |
|--------|----------|--------|
| **Docs** | `/v2/healthchecks` | v2 |
| **Provider** | `/v1/healthchecks` | v1 |

**Fix:** Update healthchecks.go line 12: change "/v1/healthchecks" to "/v2/healthchecks"

## outages

| Source | Endpoint | Version |
|--------|----------|--------|
| **Docs** | `/v2/outages` | v2 |
| **Provider** | `/v1/outages` | v1 |

**Fix:** Update outages.go line 12: change "/v1/outages" to "/v2/outages"

---

Summary: 2 mismatches, 5 matches out of 9 endpoints
