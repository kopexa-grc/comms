# Changelog

## [2.4.1](https://github.com/kopexa-grc/comms/compare/v2.4.0...v2.4.1) (2026-02-24)


### Bug Fixes

* use correct URL field name in review-overdue templates ([fcb6e1b](https://github.com/kopexa-grc/comms/commit/fcb6e1b3c65dba8b93af925f611eb052ad269421))

## [2.4.0](https://github.com/kopexa-grc/comms/compare/v2.3.0...v2.4.0) (2026-02-24)


### Features

* add data subject request receipt confirmation email ([d535265](https://github.com/kopexa-grc/comms/commit/d5352651bd99299783c3909fc19d55fcdfb6cb79))
* add incident deadline reminder and overdue email templates ([375005a](https://github.com/kopexa-grc/comms/commit/375005a8cf146c50f2d10d1a4967258147bcd295))

## [2.3.0](https://github.com/kopexa-grc/comms/compare/v2.2.0...v2.3.0) (2026-02-14)


### Features

* add organization transfer email support ([be66cf4](https://github.com/kopexa-grc/comms/commit/be66cf4879b205d0269fbc93f9fc11e41303e215))

## [2.2.0](https://github.com/kopexa-grc/comms/compare/v2.1.1...v2.2.0) (2026-02-13)


### Features

* add vendor survey OTP email template and SendSurveyOTP function ([967083e](https://github.com/kopexa-grc/comms/commit/967083ee16150f65046c80fcd8395275fb710ac2))


### Bug Fixes

* Update company address in email templates and vendor survey ([c2acd21](https://github.com/kopexa-grc/comms/commit/c2acd21baad5ceed43ce18a856b69d9f17f2c6e8))

## [2.1.1](https://github.com/kopexa-grc/comms/compare/v2.1.0...v2.1.1) (2025-10-09)


### Bug Fixes

* correct url parsing ([dce10aa](https://github.com/kopexa-grc/comms/commit/dce10aa83419a2f0732fd7834d46ec9566684bf6))

## [2.1.0](https://github.com/kopexa-grc/comms/compare/v2.0.0...v2.1.0) (2025-10-09)


### Features

* bumped to v2 ([fb5b995](https://github.com/kopexa-grc/comms/commit/fb5b9950fa9b0228702489f18213fd4ea93a7abd))

## [2.0.0](https://github.com/kopexa-grc/comms/compare/v1.2.0...v2.0.0) (2025-10-09)


### ⚠ BREAKING CHANGES

* bumped email templates to URL mode

### Features

* bumped email templates to URL mode ([48a85b3](https://github.com/kopexa-grc/comms/commit/48a85b31f89d30a05f9a604bc6f5f0bc71475377))
* recovery codes regenerated email ([c04e673](https://github.com/kopexa-grc/comms/commit/c04e67366990655f65f60ac932991cddf5527f4a))

## [1.2.0](https://github.com/kopexa-grc/comms/compare/v1.1.0...v1.2.0) (2025-09-24)


### Features

* add review overdue email functionality with templates and sending logic ([b85225b](https://github.com/kopexa-grc/comms/commit/b85225b301cb6ac687c077fdb02754a495fded49))


### Bug Fixes

* format struct fields in ReviewOverdueData for consistency ([da2f731](https://github.com/kopexa-grc/comms/commit/da2f731667a28de4532c9bd1d6f795feecf244a8))

## [1.1.0](https://github.com/kopexa-grc/comms/compare/v1.0.0...v1.1.0) (2025-07-23)


### Features

* add subscription email functionality with HTML and text templates, and implement sending logic ([dd81130](https://github.com/kopexa-grc/comms/commit/dd811309ce979bd696bb8351e16a449df1a6621f))


### Bug Fixes

* remove unnecessary blank line in SendNewSubscriber function ([2b778fb](https://github.com/kopexa-grc/comms/commit/2b778fb862ebc5153dcf1882649503d0342c9cdc))

## 1.0.0 (2025-06-06)


### Features

* add functionality for sending invite accepted emails with templates and tests ([ec1acc1](https://github.com/kopexa-grc/comms/commit/ec1acc12a383192f1b434a760997ccb4897278c8))
* implement password reset success email functionality with templates and tests ([200a374](https://github.com/kopexa-grc/comms/commit/200a37495fa187d13d33efa1aef7745c628ff93e))
* implement vendor assessment email functionality with templates and tests ([289f98e](https://github.com/kopexa-grc/comms/commit/289f98e5e444de7889bac06c878cb529eecb3ad5))
* verification email moved to comms repo ([#1](https://github.com/kopexa-grc/comms/issues/1)) ([af62bdd](https://github.com/kopexa-grc/comms/commit/af62bddac08cf7d9739bfad1901818db4941722d))
