Swan Roadmap
============

### About this document

Here we provide high-level description of desired features which are gathered from the community and planned by the Swan team. This should serve as a reference point for Swan users and contributors to understand the direction that the project is heading to, and help determined if a contribution could be conflicting with the long term plan or not.

### How to help?

Discussion on the roadmap can take place in threads under [Issues](https://github.com/Dataman-Cloud/swan/issues). Please open or comment on an issue if you want to provide suggestions and feedback to an item in the roadmap. You'd better review the roadmap before file a issue to avoid potential duplicated effort.


### Version 0.x.x Objectives, ETA `6/30/2017`

`Swan` team's main goals for version 0.x.x are to make it more stable
and production ready. Currently we scoped 3 main features as `Swan` core
functionalities, they are

  * application related features inucludes scaling, rolling-update,
    health-check.
  * `Swan` itself should be highly avaliable software before put it into
    production.
  * API Gateway help `Swan` user expose their own HTTP API to the
    public, only for `repliactes` applications.


### Version 1.x.x Objectives

more feature are suppose to be added into `Swan` if the core
functionalities are stable as we expected, of them includes

  * more features added to API Gateway, like more protocols supposed to
    be supported.
  * service discovery easily connect to Nginx or HAProxy
  * quota management & preemptive tasks.
  * more metrics reflect runtime evnronment of `Swan` itself.
