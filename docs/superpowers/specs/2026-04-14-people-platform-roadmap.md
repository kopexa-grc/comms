# People Platform — Roadmap

**Status:** Draft
**Date:** 2026-04-14
**Owner:** Julian Koehn
**Type:** Roadmap / Decomposition (this is a planning document, not an implementation spec — each block below gets its own design + plan when its turn comes)

---

## 1. Vision

Build the People-anchored compliance backbone Kopexa needs to pass HR-security audits (ISO 27001 Annex A.6, SOC 2 CC1/CC6) and operate as the single source of evidence for *who works here, what they had to do, and whether they did it*.

Today, Kopexa has a `People` schema (HR roster) and `BusinessUnit` membership, but no Joiner-Mover-Leaver (JML) history, no document acknowledgement tracking, no training records, no access reviews, no employee-facing surface, and no HR-system integration. This roadmap fills those gaps in eight building blocks.

## 2. Goals

- **Audit-ready evidence** for ISO 27001 Annex A.6 (HR security), A.5.18 (access rights), A.5.10 (acceptable use), and SOC 2 CC1.4 / CC6.2 / CC6.3.
- **Single canonical roster** of current and former employees, sourced from the customer's HR system (Personio, HR Works, generic CSV) with field-specific sync rules.
- **Historical truth**: answer "who was in department X on date Y" and "who held account Z on date Y" without ambiguity.
- **Workflow automation** for joiner / mover / leaver flows, including assignment of policies, NDAs, trainings, and access provisioning/deprovisioning tasks.
- **Self-service surface** for employees including office workers (magic link), enterprise SSO, and shop-floor workers via shared kiosks.
- **Manager-proxy fallback** for every employee-facing workflow, so the platform is operable in customers where employees cannot log in themselves.

## 3. Non-Goals (explicitly out of scope for this roadmap)

- **Full LMS** — video hosting, SCORM, content authoring, in-app questionnaires/assessments. V1 of training tracking covers external completion only.
- **Performance management** — reviews, goals, 1:1s, OKRs.
- **Payroll / time tracking** integration.
- **Mobile native app.** Workforce Hub V1 is responsive web only.
- **People as asset owners.** Asset/control/risk ownership stays on `User` (Kopexa platform user). People are the HR roster, not actor identities.

## 4. Naming / Conceptual Model

| Concept | Schema | What it represents |
|---|---|---|
| **People** | `people` | HR roster entry: every current or former employee. May or may not have a Kopexa login. Sourced from HR system. |
| **User** | `user` | Kopexa platform login. Acts in the system. May be linked to a `People` row, may not (e.g. external auditors, contractors). |
| **Account** | `account` | An identity in an external system (M365, AWS, Okta). Already linked to People via `account.people_id`. |
| **BusinessUnit** | `business_unit` | Department / team. Hierarchical. |
| **BusinessUnitMembership** | `business_unit_membership` | Existing User↔BusinessUnit join. *Stays untouched* — it tracks Kopexa users, not HR people. |

Critical separation: **`User` ≠ `People`**. This roadmap operates entirely in the People domain. Where a workflow needs an actor (a manager signing off), that actor is a `User`; where it needs a subject (the person being onboarded), that subject is a `People`.

---

## 5. Building Blocks

The eight blocks below are sequenced top-to-bottom. Each block is independently shippable (its own design + plan + implementation cycle). Dependencies are explicit in each block's "Depends on" line.

### Block 1 — People Core + HR Import

**Goal:** Make `People` the canonical HR roster, sourced from the customer's HR system with configurable per-field sync rules.

**Why first:** Everything else hangs off People. No People → no JML → no acks → no access reviews.

**Key deliverables**
- HR connector framework (`internal/services/hr_sync/` or similar). First connector: **Personio**. Second: generic **CSV upload**. Future: HR Works, BambooHR, SCIM.
- Schema deltas on `people`:
  - `external_id` (string, optional, unique per `(source, space_id)`)
  - `source` (enum: `personio`, `csv`, `manual`, …)
  - `last_synced_at` (time, optional)
  - `sync_state` (enum: `synced`, `drifted`, `pending`, `error`)
  - `sync_error` (text, optional)
- New schema: `hr_field_mapping` per `(space_id, source)` defining for each People field a sync mode:
  - `source_wins` — HR system overwrites Kopexa on every sync (default for identity fields)
  - `kopexa_wins` — HR sync ignores this field (default for GRC-only fields)
  - `manual_review` — sync surfaces a diff, admin resolves
- Sync runner as a River job, scheduled per space. Diff report + audit trail per sync run.
- Admin UI for connector setup, mapping configuration, sync history, manual sync trigger.

**Default mapping (Personio)**
- `source_wins`: `email`, `first_name`, `last_name`, `start_date`, `end_date`, `job_title`, `department`, `employment_status`
- `kopexa_wins`: anything GRC-specific added in later blocks (training status, ack state, etc.)

**Depends on:** nothing. Foundational.

---

### Block 2 — JML History (Department Membership Over Time)

**Goal:** Track Joiner-Mover-Leaver history precisely enough to answer "who was in department X on 2025-06-01" and "what was Person Y's department history."

**Why critical:** Auditors do point-in-time scoping. Today's flat `BusinessUnit ↔ People` M2M cannot answer historical questions and silently rewrites the past on every change.

**Key deliverables**
- New schema: `people_business_unit_membership`
  - `people_id`, `business_unit_id`, `started_at`, `ended_at` (nil = current), `reason` (enum: `hire`, `transfer`, `restructure`, `correction`)
  - Index on `(people_id, ended_at IS NULL)` for "current memberships" lookups
  - Index on `(business_unit_id, started_at, ended_at)` for point-in-time queries
- New schema: `people_employment_event` — append-only audit log
  - `people_id`, `event_type` (enum: `hired`, `department_changed`, `role_changed`, `terminated`, `rehired`), `effective_at`, `payload` (JSON snapshot of what changed), `actor_user_id` (or `system` for HR sync)
- Migration: collapse the existing flat `BusinessUnit ↔ People` M2M into `people_business_unit_membership` rows with `started_at = created_at` and `ended_at = NULL`. Existing data is treated as "open since first observed."
- Hooks: HR sync writes employment events; the existing flat M2M edge is removed in favor of the new history table.
- Query helpers: `MembershipsAt(people_id, t)`, `MembersOf(business_unit_id, t)`.

**Depends on:** Block 1 (HR sync writes the events; manual edits do too).

---

### Block 3 — Document Assignments via Business Unit (Policies + NDAs)

**Goal:** A document (policy, handbook, NDA) is assigned to one or more business units. Every Person currently in those units inherits the obligation. Acknowledgements are tracked per Person × Document version, with full evidence trail.

**Why this shape:** Per-person assignment doesn't scale. Per-BU assignment matches how compliance is actually managed ("everyone in Engineering signs the secure-coding policy"). NDAs are simply documents flagged `requires_signature=true`.

**Key deliverables**
- New schema: `document_assignment`
  - `document_id`, `business_unit_id`, `requires_signature` (bool), `version_pinning` (enum: `latest` — re-ack on new version, or `pinned` — fixed version), `due_within_days` (int, optional — for new joiners)
- New schema: `document_acknowledgement`
  - `people_id`, `document_id`, `document_version_id`, `acked_at`, `methodology` (enum: `self_service_web`, `kiosk`, `manager_proxy`, `paper_upload`, `imported`), `actor_user_id` (who recorded it; may equal the person's linked User or be a manager), `evidence_file_id` (optional, e.g. scanned signature)
  - Unique index on `(people_id, document_id, document_version_id)`
- Resolver: given a Person, return all documents the person currently owes (assignment via membership × version-pinning policy × ack history).
- Hook: when a Person joins a BU (via Block 2 event), a `document_obligation` is materialized for them. When they leave, obligations remain in history but are no longer "open."
- **Manager-proxy UI in the Main app** (mandatory in this block): admin/manager view per Person showing open obligations, with "sign as proxy" / "upload paper signature" actions. Without this, the block is unusable for customers without Workforce Hub.
- Audit-trail invariants: every ack stores `actor_user_id`, `methodology`, optional `evidence_file_id`. A person signing themselves through Workforce Hub looks identical in the audit log to a manager signing as proxy, except for the `actor_user_id` ≠ linked User of the Person and the `methodology` field.

**Depends on:** Blocks 1 + 2.

---

### Block 4 — Training Catalog + Assignments

**Goal:** Track that the right people completed the right training within the right period. V1 covers tracking only — actual learning happens externally.

**Out of scope for V1:** video hosting, in-app quizzes, SCORM, content authoring. This is *not* an LMS. It is an *evidence layer* for trainings the customer runs elsewhere (or holds in person).

**Key deliverables**
- New schema: `training`
  - `name`, `description`, `category` (e.g. `security_awareness`, `gdpr`, `code_of_conduct`), `external_url` (optional — points to the actual training in a third-party LMS), `default_recurrence` (enum: `once`, `yearly`, `every_2_years`)
- New schema: `training_assignment`
  - `training_id`, `business_unit_id`, `recurrence` (overrides default), `due_within_days` (for joiners), `grace_period_days`
- New schema: `training_completion`
  - `people_id`, `training_id`, `period_start`, `period_end` (a yearly training in 2025 has period 2025-01-01 to 2025-12-31), `completed_at`, `methodology` (enum same as ack), `actor_user_id`, `evidence_file_id` (optional certificate)
  - Unique index on `(people_id, training_id, period_start)`
- Resolver: open training obligations per Person.
- **Manager-proxy UI** in Main app (mandatory): mark training as completed for a Person, attach evidence.

**Depends on:** Blocks 1 + 2. Parallelizable with Block 3.

---

### Block 5 — Onboarding / Offboarding Workflows

**Goal:** When a JML event happens (hire, terminate, department change), automatically materialize a checklist of tasks: doc-acks to do, trainings to assign, asset returns to confirm, account provisioning/deprovisioning to request. Track each item to completion.

**Key deliverables**
- New schema: `workflow_template`
  - `space_id`, `trigger_event` (enum mapped to `people_employment_event.event_type`), `name`, `version`
- New schema: `workflow_template_step`
  - `workflow_template_id`, `step_type` (enum: `assign_documents`, `assign_trainings`, `request_account_provisioning`, `request_account_deprovisioning`, `return_asset`, `manual_task`), `order`, `assignee_role` (enum: `manager`, `it`, `hr`, `security`), `due_within_days`, `config` (JSON for step-specific params)
- New schema: `workflow_run` — one per Person × triggering event
  - `people_id`, `workflow_template_id`, `triggering_event_id`, `started_at`, `completed_at`, `status` (enum: `running`, `completed`, `cancelled`)
- New schema: `workflow_task` — one per step instance
  - `workflow_run_id`, `template_step_id`, `assignee_user_id` (resolved from `assignee_role`), `due_at`, `status`, `completed_at`, `completed_by_user_id`, `notes`
- Hook on `people_employment_event` insert: look up matching templates, instantiate runs.
- Resolution helpers: `assign_documents` step writes obligations using Block 3 resolver; `assign_trainings` uses Block 4. `request_account_*` opens a task for IT — actual provisioning is out of scope (manual handoff).
- **Main-app UI**: workflow runs per Person, task inbox per assignee role, escalation on overdue.

**Depends on:** Blocks 1 + 2 + 3 + 4.

---

### Block 6 — Access Reviews

**Goal:** Periodic review campaigns: pick a scope (BU, system, role), pull current Accounts linked to People in scope, route each Account to a reviewer, capture decisions (keep / revoke / change / needs_info), generate evidence.

**Key deliverables**
- New schema: `access_review_campaign`
  - `space_id`, `name`, `scope` (JSON: `business_unit_ids`, `account_sources`, `role_filter`), `period_start`, `period_end`, `reviewer_user_id` (or `assignee_strategy`: `manager_of_person`, `business_unit_owner`), `status` (enum: `draft`, `running`, `completed`, `cancelled`)
- New schema: `access_review_item`
  - `campaign_id`, `account_id`, `people_id` (resolved from `account.people_id`), `reviewer_user_id`, `decision` (enum: `keep`, `revoke`, `change`, `needs_info`, `pending`), `reason`, `decided_at`
- Generator: at campaign launch, materialize one item per in-scope Account. Reviewer routing uses the selected strategy.
- River job for periodic auto-creation of campaigns (e.g. quarterly access review for production systems).
- Evidence pack export per campaign: items + decisions + reviewer + timestamp → PDF/XLSX.
- **Main-app UI**: reviewer task list, campaign overview, decision interface.

**Depends on:** Block 1 (People) + existing `account.people_id` linkage. Independent of Blocks 3-5.

---

### Block 7 — Compliance Reporting / Evidence Packs

**Goal:** Pre-built audit-ready reports per ISO control. Read-only views on Blocks 1-6. No new data model.

**Key deliverables**
- Report definitions, each rendered as PDF + XLSX:
  - **A.6.1 — Screening / NDA evidence**: per-Person NDA acknowledgement status with version, date, methodology
  - **A.6.3 — Training & awareness**: completion rate by training × period × BU; list of overdue people
  - **A.6.5 — Termination**: offboarding workflow completion rate; list of incomplete offboardings
  - **A.5.18 — Access rights**: latest access review per account/system; list of accounts not reviewed in current period
  - **A.5.10 — Acceptable use**: acknowledgement status of acceptable-use policy per active employee
  - **People register**: full roster with employment status, department, dates (HR snapshot)
- Each report parameterizable by `space_id`, `as_of` date (point-in-time using Block 2 history), and optional BU filter.
- Export endpoints follow the SoA-export pattern (REST, requires X-Space-ID header, blob download).
- Dashboard tile in Main app: traffic-light per report.

**Depends on:** Blocks 1-6.

---

### Block 8 — Workforce Hub *(separate host)*

**Goal:** Self-service portal where employees themselves complete their compliance tasks — and where shop-floor workers access operational documentation via shared kiosks. Separate host, branded per organization, multi-modal authentication.

**URL strategy:** V1 = `<orgslug>.employee.kopexa.com` (subdomain — clean cookie/CORS isolation, mirrors Trust Center). V2 = custom domains.

**Three modules sharing one auth + UX shell:**

#### 8a — Compliance Self-Service
The employee's view of their own obligations: pending document acknowledgements, due trainings, open onboarding/offboarding tasks. Uses Block 3 / 4 / 5 backends. Writes acks/completions with `methodology = self_service_web` (or `= kiosk` if running inside the kiosk shell).

#### 8b — Knowledge Reader
Browse and search published operational documentation: SOPs, work instructions, processes, policies — read-only. Per-document read-receipts capture proof of access via a new `document_read_receipt` schema (`people_id`, `document_version_id`, `read_at`, `methodology` from the audit-trail provenance bundle). Bridges Block 3 documents into operational use.

#### 8c — Kiosk Shell *(shared capability)*
A UX shell + auth provider used by both 8a and 8b. Optimized for shared touch terminals on the shop floor:
- Auth: badge scan / employee number + PIN
- Large touch UI, no browser chrome
- Auto-logout on inactivity (configurable, default 60s)
- Per-interaction audit trail
- Optional terminal pinning per BU (a kiosk in production hall A defaults its content to that hall's documents)

**Authentication modes (apply to 8a + 8b):**
| Mode | Use case | Notes |
|---|---|---|
| **Magic link via email** | Office workforce | Already implemented for Trust Center — reuse |
| **SSO (SAML / OIDC)** | Enterprise customers with IdP | Map by email or `external_id` to People |
| **Kiosk (badge + PIN)** | Shop floor, shared terminals | New auth provider; PIN reset by manager |

**Manager-proxy stays in Main app**, not in Workforce Hub. WFH is for the *subject* (the People row) to act on themselves.

**Audit-trail invariants:**
- Every action stores `methodology` (`self_service_web` / `kiosk` / `sso` / `magic_link`) plus the originating IP / device fingerprint where available
- A self-served Workforce Hub ack and a manager-proxy ack from Main app are *both* valid evidence; the auditor can filter by methodology to see how each ack was produced
- A Person signing through SSO shows `actor_user_id` = the User row created/linked at SSO time

**Depends on:** Blocks 1 + 3 + 4 (data) and Block 5 (tasks). Block 8 can start design once Block 3 is shipped, but UI cannot be useful until 4 + 5 are also shipped.

---

## 6. Cross-Cutting Concerns

### 6.1 Audit-Trail Invariants

Every write operation on Blocks 3-6 stores a uniform **provenance bundle**:
- `actor_user_id` — the Kopexa User who performed the action (may be the Person's linked User, may be a manager, may be the system for HR sync)
- `subject_people_id` — the Person the action concerns
- `methodology` — how the action was performed: `self_service_web`, `kiosk`, `sso`, `magic_link`, `manager_proxy`, `paper_upload`, `imported`
- `evidence_file_id` — optional pointer to a file (signature scan, certificate)
- `client_meta` — optional IP / user agent / device

This bundle should live as a **mixin** so it is consistent across `document_acknowledgement`, `training_completion`, `access_review_item.decision`, `workflow_task.completion`, and `document_read_receipt`.

### 6.2 Privacy & Authorization

People records contain sensitive employment data and must be tightly access-controlled:
- **Existing privacy policies** on `people` continue to apply (`AlwaysAllowRule` query + standard mutation policies). Verify these still satisfy the new use cases.
- **FGA-tuples** to add:
  - `space → can_view_people: hr_role | security_role | manager_of_person`
  - `business_unit → can_view_members: business_unit_owner | hr_role`
- **People themselves**: when a Person has a linked User, that User must be able to view *only their own* People row plus their own obligations. Achieved via interceptor on Workforce Hub queries (similar to the Notification user-scope interceptor).
- **HR sync** runs as a system actor with elevated read but write-only-via-mapping, never bypasses privacy on adjacent entities.

### 6.3 Manager-Proxy as the Universal Fallback

Blocks 3, 4, 5 (and indirectly 6, 7) **must** be operable through the Main app without Workforce Hub. The Manager-Proxy UI is non-negotiable in those blocks because:
- Some customers will never deploy Workforce Hub (cost / complexity / culture)
- Even with Workforce Hub deployed, edge cases (sick leave, paper signatures, corrections) require admin intervention
- Workforce Hub is an *optimization*, not a *prerequisite*

### 6.4 i18n

All employee-facing surfaces (Workforce Hub, emails, kiosk shell) must support German *(du-form, never Sie)* and English from day one. Re-uses the shared `internal/i18n` bundle and `resources/locales/{de,en}.yaml` already established for emails.

### 6.5 Migration Strategy

The existing flat `BusinessUnit ↔ People` M2M (Block 2) is the only hard migration. All other blocks are additive — new schemas, new endpoints, no breaking changes to current entities. Block 2's migration writes one open membership row per existing edge and drops the edge table; backfill is fast because the volume is small (people are O(thousands), not O(millions)).

---

## 7. Sequencing & Dependencies

```
Block 1 (People + HR Sync)
    │
    ▼
Block 2 (JML History)
    │
    ├──────────────┐
    ▼              ▼
Block 3        Block 4
(Doc Acks)    (Trainings)
    │              │
    └──────┬───────┘
           ▼
       Block 5 (Workflows)
           │
           ▼
       Block 6 (Access Reviews)   *(can also start in parallel after Block 1)*
           │
           ▼
       Block 7 (Reporting)
           │
           ▼
       Block 8 (Workforce Hub)
```

**Recommended delivery order:** 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8.

**Parallelization opportunities:**
- Block 4 can start as soon as Block 2 lands (parallel with Block 3)
- Block 6 only needs Block 1 and the existing `account.people_id` link — could ship earlier if priorities shift
- Block 8 design work (auth modes, kiosk shell, micro-frontend bootstrapping) can begin once Block 3's data contract is stable, even if Block 5 isn't done yet

---

## 8. Open Questions (defer to per-block specs)

These are intentionally not answered here; each gets resolved in its own spec.

- **Block 1**: Webhook vs polling for Personio? Initial sync transactional or streamed?
- **Block 2**: How are corrections handled — soft-delete the wrong membership and create a corrected one, or update in place with audit log?
- **Block 3**: When a document version changes mid-period, do all existing acks stay valid (pinned semantics) or invalidate immediately (latest semantics)? Probably configurable per document.
- **Block 4**: How does period rollover work for yearly trainings — automatic on Jan 1 or aligned to the customer's compliance year?
- **Block 5**: Templates per-space or also templates inheritable from a parent organization?
- **Block 6**: Reviewer can delegate to another reviewer — yes or no?
- **Block 7**: Do reports render server-side (chromedp/wkhtmltopdf) or use the existing XLSX export pattern (excelize)?
- **Block 8**: Kiosk auth — badge format (RFID UID, NFC, barcode)? PIN length? Lockout policy?

---

## 9. Success Criteria for the Full Roadmap

The People Platform is "done" (V1 complete) when a Kopexa customer can:

1. **Connect Personio**, see their full employee roster auto-populate, and have new hires appear within one sync interval.
2. **Define an onboarding workflow** that, when a new employee starts, automatically: assigns the IT-security policy + acceptable-use policy + NDA, schedules security awareness training, opens a task for IT to provision M365, and opens a task for the manager to confirm asset handover.
3. **See in a single Main-app view** for any Person: current department, employment events history, all open obligations, all completed evidence with timestamps and methodology.
4. **Run a quarterly access review** of M365 admin accounts, route each item to the appropriate manager, and export the resulting decisions as an evidence pack for the auditor.
5. **Deploy Workforce Hub** at `<orgslug>.employee.kopexa.com`, have office workers self-serve via magic link, and have shop-floor workers access policies + complete trainings via a kiosk terminal logged in by badge.
6. **Print the A.6.3 training-completion report** for any historical date and prove that 100% of in-scope employees completed mandatory training within the period.

---

## 10. What This Document Is Not

This is a **roadmap**, not an implementation spec. Each of the eight blocks gets:

1. Its own design document (using the brainstorming → spec flow)
2. Its own implementation plan (using the writing-plans flow)
3. Its own implementation, review, and ship cycle

This document exists to make sure we agree on **what we are building**, **why we are building it**, **in which order**, and **where the boundaries are** — before we sink time into any one piece.
