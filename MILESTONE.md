MILESTONE
==============

Here we list series of milestones for both `Swan` users and `Swan` team
to follow, we will try our best to make sure everything goes with the
milestones we set ahead.

### \#1 feature improvement & `Swan` stabilization 1 (2-17-2017)

  - **[IMPROVEMENT]** `Constraints` improvement
  - **[FEATURE]**      Agent can be removed out from the cluster
  - **[BUG]**          DNS record issue for fixed type application
  - **[IMPROVEMENT]** `./bin/swan` default without any arguments should display help
  - **[IMPROVEMENT]**  add help API to return all API access points of Swan
  - **[IMPROVEMENT]**  separate irrelevant components out of swan codebase. eg. swan cli, go swan, swan frontend(make it standalone swan-ui project)



### \#2 feature improvement & `Swan` stabilization 2 (2-24-2017)

  - **[IMPROVE]** rename appID to appName in version JSON, import
    appName which is the name of an app.
  - **[IMPROVE]** make sure fixed type app allow customized their
    network name.

  - **[IMPROVE]** Handle recind offer
  - **[IMPROVE]** Handle inverse offer
  - **[IMPROVE]** make sure framework handle mesos leader shift gracefully
  - **[IMPROVE]** make sure framework handle agent lost gracefully
  
  - **[FEATURE]** add node status and heartbeats from leader to agent

### \#3 feature improvement & `Swan` stabilization 3 (3-3-2017)

  - **[BUG]** task IP lost after `Swan` restart
  - **[BUG]** DNS A record lost after `Swan` restart for fixed app
  - **[IMPROVE]** Handle recind offer
  - **[IMPROVE]** Handle inverse offer
  - **[FEATURE]** restart policy revisit
  - **[IMPROVE]** remove mixed mode
  - **[IMPROVE]** supprt host mode network for replicates app


### \#4 feature improvement & `Swan` stabilization 4 (3-10-2017)

  - **[IMPROVE]** rolling update, deletion, scale should not in batch
    but one after another

  - **[IMPROVE]** revisit janitor
  - **[IMPROVE]** revisit dns


### \#5 feature improvement & stabilization phase 5 (3-20-2017)

  - **[Testing]** increase unit testing cover rate up to 50%
  - **[HA]** test HA mode in production alike environment
  - **[Desgin]** design for service/app group feature
  
