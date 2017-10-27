v1.1.0 (2017-10-27)
-------------------
* Add ability to publish and receive binary data. Check README and examples
 to get more information;
* Add support of HTTPS proxy;
* Get rid of CODE_DATA_REQUEST response code.

v1.0.1 (2017-07-07)
-------------------
* Catch *panic* in user callbacks. Extend subscription listener by "OnPanicRecover" action;
* Close connection before fire the Stop callback;
* FSM: Add RWMutex when changing/getting current state to avoid race condition;
* RTM Client: Fix race condition when subscribing;
* Fix panic when trying to print "RTM Error" with nil error;
* Change code for ERROR_CODE_AUTHENTICATION const;
* Rename RTM struct to RTMClient;
* Add examples.

v0.2.0 (2017-04-28)
-------------------
* New Subscription model **[no-backward-compatibility]**:
  - Get rid of data channel;
  - Add Listener instead of Observer. All callbacks must be specified before
  subscription is created. subscription.On/Once no longer work;
  - Change subscription.New() signature;
* Add Event helpers for RTM Client.   

v0.1.0 (2017-04-21)
-------------------
* Initial release
