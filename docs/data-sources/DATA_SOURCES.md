# Data Sources Catalog

> No public source says "there is a free spot here, now". Open data provides **priors** (where parking is allowed, capacity, zone saturation). Live Spots come only from drivers.

| Source | Provides | Coverage | Used as | Refresh | MVP |
|---|---|---|---|---|---|
| **Own passive signals** (parking / departure detection) | Real rotation per street & hour | Wherever users are | Priors (strongest), Spot outcomes | Real time | ✅ |
| **OpenStreetMap** street parking tags (`parking:*`, `parking=lane/street_side`) | Parkable segments, orientation, side | Spain (uneven) | `parkable_segments` | Weekly | ✅ |
| **Madrid SER** datasets (zones, streets & number of bays, parking meters) | Regulated zones, capacity | Madrid city | `regulated_zones`, capacity | Monthly | ✅ |
| **Madrid traffic intensity** | Traffic load per measurement point (~5 min) | Madrid city | Saturation proxy in priors | 5–15 min | ✅ (batch) |
| **DGT National Access Point (NAP)** — DATEX II | Camera locations & images, road data | Spain except Basque Country & Catalonia | Camera catalog for F2+, other cities | Varies | Catalog only |
| **Madrid Informo cameras** | Still images every ~5 min | Madrid city | Zone occupancy estimate (Urban Vision) | 5 min | ✅ flag OFF |
| **Public car park occupancy** (various cities, e.g., Málaga) | Real-time occupancy of off-street car parks | City-dependent | Zone saturation signal | Real time | ❌ (F3) |
| **Other municipalities' ORA zones** | Regulated areas | Per municipality | `regulated_zones` | Manual / OSM | Via OSM + crowd |

## Adapter contract

Every source is implemented as an adapter in `services/ingestion/sources/<source>/` (or `services/vision/cameras/<provider>/`) and must declare:

- `source_id`, licence and attribution requirements
- coverage geometry
- output target table and mapping
- refresh cadence
- data-quality checks (row counts, geometry validity, freshness)

Adapters are enabled per Coverage area in `coverage_areas.enabled_sources`. Adding a city = new adapters (if needed) + configuration. Core code does not change.

## Licensing notes

- OSM data: ODbL — attribution required, share-alike on derived databases.
- Madrid open data: municipal licence with attribution clause.
- DGT NAP: check licence per dataset (several are Creative Commons).
- Cameras: images are processed transiently and never redistributed.
