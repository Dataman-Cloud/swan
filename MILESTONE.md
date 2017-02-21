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
  - **[FEATURE]** restart policy revisit.

  - **[IMPROVE]** Handle recind offer
  - **[IMPROVE]** Handle inverse offer
  - **[IMPROVE]** make sure framework handle mesos leader shift gracefully
  - **[IMPROVE]** make sure framework handle agent lost gracefully
  
  - **[FEATURE]** add node status and heartbeats from leader to agent
