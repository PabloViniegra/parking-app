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

## Coverage boundary

The Comunidad de Madrid Coverage boundary comes from the **Instituto Geográfico Nacional de España (IGN)**, Base de Datos de Divisiones Administrativas de España, distributed by the **Centro Nacional de Información Geográfica (CNIG)**. The [dataset metadata](https://www.idee.es/csw-codsi-idee/srv/api/records/spaignLLM) specifies CC BY 4.0; see also the [IGN data policy](https://centrodedescargas.cnig.es/CentroDescargas/politica-datos). Attribution: © Instituto Geográfico Nacional de España / Centro Nacional de Información Geográfica.

The [administrative-unit feature](https://api-features.ign.es/collections/administrativeunit/items/1174523?f=json), national code `34130000000`, was retrieved on 2026-11-04 (UTC). SHA-256 of the original response: `d87f2d69597e2f006df241e2b3d40f269f50f6bc8187803bb2dc151d79a43be9`. The migration stores its unmodified geometry as longitude/latitude in SRID 4326; updates require a new reviewed migration. The IGN notes that boundaries generally have an uncertainty of around 40 m and some segments may be provisional. Coverage decisions near the boundary therefore have the same limitations.
