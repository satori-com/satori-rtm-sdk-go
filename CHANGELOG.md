v0.2.0
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
