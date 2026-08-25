package plugin

import (
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// This type saves a wrapped_json type of Thruk response
// { "data": [] , "meta": [] }
type ThrukWrappedJsonResponse struct {
	Data []map[string]any              `json:"data"`
	Meta *ThrukWrappedJsonResponseMeta `json:"meta"`
}

// This type saves "meta" object in wrapped_json type of Thruk response
// RequestDuration: is added later, not present in Thruk Response
// ParseDuration: is added later, not present in Thruk response
type ThrukWrappedJsonResponseMeta struct {
	Columns         []ThrukWrappedJsonResponseMetaColumn `json:"columns"`
	RequestDuration time.Duration                        `json:"requestDuration"`
	ParseDuration   time.Duration                        `json:"parseDuration"`
}

// This type saves elements of "meta"."columns" array in wrapped_json type of Thruk responses
// Most of the colum metadata only have "name"
// Some might have "type" as well, taking values like: "time"
// Some might have "config" which is a nested object like: { "unit" : "s"},
// GrafanaDataType: added later, not present in Thruk Response. It serves to save the parsed Grafana SDK type in the same struct
type ThrukWrappedJsonResponseMetaColumn struct {
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	GrafanaDataType data.FieldType `json:"grafanaDataType"`
	Config          any            `json:"config"`
}

// ThrukAPIEndpoint represents all available REST API endpoints of Thruk.
// Generated from Thruk's indexer JSON; type is int for easy comparison.
// the indexer response can be found in docs/r-v1-index-response.json
// Enum values are auto-assigned via iota starting at 0.
type ThrukAPIEndpoint int

const (
	// EndpointListRoot GET / lists all available rest urls.
	EndpointListRoot ThrukAPIEndpoint = iota
	// EndpointListAlerts GET /alerts lists alerts based on logfiles.
	EndpointListAlerts
	// EndpointListCheckStats GET /checks/stats lists host / service check statistics.
	EndpointListCheckStats
	// EndpointSendCommand POST /cmd Sends any command.
	EndpointSendCommand
	// EndpointListCommands GET /commands lists livestatus commands.
	EndpointListCommands
	// EndpointListCommandByName GET /commands/<name> lists commands for given name.
	EndpointListCommandByName
	// EndpointGetCommandConfig GET /commands/<name>/config Returns configuration for given command.
	EndpointGetCommandConfig
	// EndpointReplaceCommandConfig POST /commands/<name>/config Replace command configuration completely, use PATCH to only update specific attributes.
	EndpointReplaceCommandConfig
	// EndpointPatchCommandConfig PATCH /commands/<name>/config Update command configuration partially.
	EndpointPatchCommandConfig
	// EndpointDeleteCommandConfig DELETE /commands/<name>/config Deletes given command from configuration.
	EndpointDeleteCommandConfig
	// EndpointListComments GET /comments lists livestatus comments.
	EndpointListComments
	// EndpointListCommentByID GET /comments/<id> lists comments for given id.
	EndpointListCommentByID
	// EndpointRunConfigCheck POST /config/check Returns result from config check. This check does require changes to be saved to disk before running the check.
	EndpointRunConfigCheck
	// EndpointGetConfigDiff GET /config/diff
	EndpointGetConfigDiff
	// EndpointDiscardConfigChanges POST /config/discard Reverts stashed configuration changes.
	EndpointDiscardConfigChanges
	// EndpointListConfigFiles GET /config/files returns all config files
	EndpointListConfigFiles
	// EndpointListFullObjects GET /config/fullobjects Returns list of all objects with templates expanded.
	EndpointListFullObjects
	// EndpointListObjects GET /config/objects Returns list of all objects with their raw config.
	EndpointListObjects
	// EndpointCreateObject POST /config/objects Create new object. Besides the actual object config, requires
	EndpointCreateObject
	// EndpointPatchObjects PATCH /config/objects Change attributes for all matching objects.
	EndpointPatchObjects
	// EndpointDeleteObjects DELETE /config/objects Delete objects based on filters.
	EndpointDeleteObjects
	// EndpointReplaceObjectByID POST /config/objects/<id> Replace object configuration completely.
	EndpointReplaceObjectByID
	// EndpointPatchObjectByID PATCH /config/objects/<id> Update object configuration partially.
	EndpointPatchObjectByID
	// EndpointDeleteObjectByID DELETE /config/objects/<id> Remove given object from configuration.
	EndpointDeleteObjectByID
	// EndpointRunConfigPrecheck GET /config/precheck Returns result from Thruks config precheck. The precheck does not require changes to be saved to disk before running the check.
	EndpointRunConfigPrecheck
	// EndpointReloadConfig POST /config/reload Reloads configuration with the configured reload command.
	EndpointReloadConfig
	// EndpointRevertConfigChanges POST /config/revert Reverts stashed configuration changes.
	EndpointRevertConfigChanges
	// EndpointSaveConfigChanges POST /config/save Saves stashed configuration changes to disk.
	EndpointSaveConfigChanges
	// EndpointListContactGroups GET /contactgroups lists livestatus contactgroups.
	EndpointListContactGroups
	// EndpointListContactGroupByName GET /contactgroups/<name> lists contactgroups for given name.
	EndpointListContactGroupByName
	// EndpointDisableContactGroupHostNotifications POST /contactgroups/<name>/cmd/disable_contactgroup_host_notifications Disables host notifications for all contacts in a particular contactgroup.
	EndpointDisableContactGroupHostNotifications
	// EndpointDisableContactGroupSVCNotifications POST /contactgroups/<name>/cmd/disable_contactgroup_svc_notifications Disables service notifications for all contacts in a particular contactgroup.
	EndpointDisableContactGroupSVCNotifications
	// EndpointEnableContactGroupHostNotifications POST /contactgroups/<name>/cmd/enable_contactgroup_host_notifications Enables host notifications for all contacts in a particular contactgroup.
	EndpointEnableContactGroupHostNotifications
	// EndpointEnableContactGroupSVCNotifications POST /contactgroups/<name>/cmd/enable_contactgroup_svc_notifications Enables service notifications for all contacts in a particular contactgroup.
	EndpointEnableContactGroupSVCNotifications
	// EndpointGetContactGroupConfig GET /contactgroups/<name>/config Returns configuration for given contactgroup.
	EndpointGetContactGroupConfig
	// EndpointReplaceContactGroupConfig POST /contactgroups/<name>/config Replace contactgroup configuration completely, use PATCH to only update specific attributes.
	EndpointReplaceContactGroupConfig
	// EndpointPatchContactGroupConfig PATCH /contactgroups/<name>/config Update contactgroup configuration partially.
	EndpointPatchContactGroupConfig
	// EndpointDeleteContactGroupConfig DELETE /contactgroups/<name>/config Deletes given contactgroup from configuration.
	EndpointDeleteContactGroupConfig
	// EndpointListContacts GET /contacts lists livestatus contacts.
	EndpointListContacts
	// EndpointListContactByName GET /contacts/<name> lists contacts for given name.
	EndpointListContactByName
	// EndpointChangeContactHostNotificationTimeperiod POST /contacts/<name>/cmd/change_contact_host_notification_timeperiod Changes the host notification timeperiod for a particular contact to what is specified by the 'notification_timeperiod' option. The 'notification_timeperiod' option should be the short name of the timeperiod that is to be used as the contact's host notification timeperiod. The timeperiod must have been configured in Naemon before it was last (re)started.
	EndpointChangeContactHostNotificationTimeperiod
	// EndpointChangeContactSVCNotificationTimeperiod POST /contacts/<name>/cmd/change_contact_svc_notification_timeperiod Changes the service notification timeperiod for a particular contact to what is specified by the 'notification_timeperiod' option. The 'notification_timeperiod' option should be the short name of the timeperiod that is to be used as the contact's service notification timeperiod. The timeperiod must have been configured in Naemon before it was last (re)started.
	EndpointChangeContactSVCNotificationTimeperiod
	// EndpointChangeCustomContactVar POST /contacts/<name>/cmd/change_custom_contact_var Changes the value of a custom contact variable.
	EndpointChangeCustomContactVar
	// EndpointDisableContactHostNotifications POST /contacts/<name>/cmd/disable_contact_host_notifications Disables host notifications for a particular contact.
	EndpointDisableContactHostNotifications
	// EndpointDisableContactSVCNotifications POST /contacts/<name>/cmd/disable_contact_svc_notifications Disables service notifications for a particular contact.
	EndpointDisableContactSVCNotifications
	// EndpointEnableContactHostNotifications POST /contacts/<name>/cmd/enable_contact_host_notifications Enables host notifications for a particular contact.
	EndpointEnableContactHostNotifications
	// EndpointEnableContactSVCNotifications POST /contacts/<name>/cmd/enable_contact_svc_notifications Disables service notifications for a particular contact.
	EndpointEnableContactSVCNotifications
	// EndpointGetContactConfig GET /contacts/<name>/config Returns configuration for given contact.
	EndpointGetContactConfig
	// EndpointReplaceContactConfig POST /contacts/<name>/config Replace contact configuration completely, use PATCH to only update specific attributes.
	EndpointReplaceContactConfig
	// EndpointPatchContactConfig PATCH /contacts/<name>/config Update contact configuration partially.
	EndpointPatchContactConfig
	// EndpointDeleteContactConfig DELETE /contacts/<name>/config Deletes given contact from configuration.
	EndpointDeleteContactConfig
	// EndpointGetContactTotals GET /contacts/totals hash of livestatus contacts totals statistics.
	EndpointGetContactTotals
	// EndpointListDowntimes GET /downtimes lists livestatus downtimes.
	EndpointListDowntimes
	// EndpointListDowntimeByID GET /downtimes/<id> lists downtimes for given id.
	EndpointListDowntimeByID
	// EndpointListHostGroups GET /hostgroups lists livestatus hostgroups.
	EndpointListHostGroups
	// EndpointListHostGroupByName GET /hostgroups/<name> lists hostgroups for given name.
	EndpointListHostGroupByName
	// EndpointGetHostGroupAvailability GET /hostgroups/<name>/availability list availability for this hostgroup.
	EndpointGetHostGroupAvailability
	// EndpointDisableHostGroupHostChecks POST /hostgroups/<name>/cmd/disable_hostgroup_host_checks Sends the DISABLE_HOSTGROUP_HOST_CHECKS command.
	EndpointDisableHostGroupHostChecks
	// EndpointDisableHostGroupHostNotifications POST /hostgroups/<name>/cmd/disable_hostgroup_host_notifications Sends the DISABLE_HOSTGROUP_HOST_NOTIFICATIONS command.
	EndpointDisableHostGroupHostNotifications
	// EndpointDisableHostGroupPassiveHostChecks POST /hostgroups/<name>/cmd/disable_hostgroup_passive_host_checks Disables passive checks for all hosts in a particular hostgroup.
	EndpointDisableHostGroupPassiveHostChecks
	// EndpointDisableHostGroupPassiveSVCChecks POST /hostgroups/<name>/cmd/disable_hostgroup_passive_svc_checks Disables passive checks for all services associated with hosts in a particular hostgroup.
	EndpointDisableHostGroupPassiveSVCChecks
	// EndpointDisableHostGroupSVCChecks POST /hostgroups/<name>/cmd/disable_hostgroup_svc_checks Sends the DISABLE_HOSTGROUP_SVC_CHECKS command.
	EndpointDisableHostGroupSVCChecks
	// EndpointDisableHostGroupSVCNotifications POST /hostgroups/<name>/cmd/disable_hostgroup_svc_notifications Sends the DISABLE_HOSTGROUP_SVC_NOTIFICATIONS command.
	EndpointDisableHostGroupSVCNotifications
	// EndpointEnableHostGroupHostChecks POST /hostgroups/<name>/cmd/enable_hostgroup_host_checks Sends the ENABLE_HOSTGROUP_HOST_CHECKS command.
	EndpointEnableHostGroupHostChecks
	// EndpointEnableHostGroupHostNotifications POST /hostgroups/<name>/cmd/enable_hostgroup_host_notifications Sends the ENABLE_HOSTGROUP_HOST_NOTIFICATIONS command.
	EndpointEnableHostGroupHostNotifications
	// EndpointEnableHostGroupPassiveHostChecks POST /hostgroups/<name>/cmd/enable_hostgroup_passive_host_checks Enables passive checks for all hosts in a particular hostgroup.
	EndpointEnableHostGroupPassiveHostChecks
	// EndpointEnableHostGroupPassiveSVCChecks POST /hostgroups/<name>/cmd/enable_hostgroup_passive_svc_checks Enables passive checks for all services associated with hosts in a particular hostgroup.
	EndpointEnableHostGroupPassiveSVCChecks
	// EndpointEnableHostGroupSVCChecks POST /hostgroups/<name>/cmd/enable_hostgroup_svc_checks Sends the ENABLE_HOSTGROUP_SVC_CHECKS command.
	EndpointEnableHostGroupSVCChecks
	// EndpointEnableHostGroupSVCNotifications POST /hostgroups/<name>/cmd/enable_hostgroup_svc_notifications Sends the ENABLE_HOSTGROUP_SVC_NOTIFICATIONS command.
	EndpointEnableHostGroupSVCNotifications
	// EndpointScheduleHostGroupHostDowntime POST /hostgroups/<name>/cmd/schedule_hostgroup_host_downtime Sends the SCHEDULE_HOSTGROUP_HOST_DOWNTIME command.
	EndpointScheduleHostGroupHostDowntime
	// EndpointScheduleHostGroupSVCdowntime POST /hostgroups/<name>/cmd/schedule_hostgroup_svc_downtime Sends the SCHEDULE_HOSTGROUP_SVC_DOWNTIME command.
	EndpointScheduleHostGroupSVCdowntime
	// EndpointGetHostGroupConfig GET /hostgroups/<name>/config Returns configuration for given hostgroup.
	EndpointGetHostGroupConfig
	// EndpointReplaceHostGroupConfig POST /hostgroups/<name>/config Replace hostgroups configuration completely, use PATCH to only update specific attributes.
	EndpointReplaceHostGroupConfig
	// EndpointPatchHostGroupConfig PATCH /hostgroups/<name>/config Update hostgroup configuration partially.
	EndpointPatchHostGroupConfig
	// EndpointDeleteHostGroupConfig DELETE /hostgroups/<name>/config Deletes given hostgroup from configuration.
	EndpointDeleteHostGroupConfig
	// EndpointGetHostGroupOutages GET /hostgroups/<name>/outages list of outages for this hostgroup.
	EndpointGetHostGroupOutages
	// EndpointGetHostGroupStats GET /hostgroups/<name>/stats hash of livestatus hostgroup statistics.
	EndpointGetHostGroupStats
	// EndpointListHosts GET /hosts lists livestatus hosts.
	EndpointListHosts
	// EndpointListHostByName GET /hosts/<name> lists hosts for given name.
	EndpointListHostByName
	// EndpointListHostAlerts GET /hosts/<name>/alerts lists alerts for given host.
	EndpointListHostAlerts
	// EndpointGetHostAvailability GET /hosts/<name>/availability list availability for this host.
	EndpointGetHostAvailability
	// EndpointAcknowledgeHostProblem POST /hosts/<name>/cmd/acknowledge_host_problem Sends the ACKNOWLEDGE_HOST_PROBLEM command.
	EndpointAcknowledgeHostProblem
	// EndpointAcknowledgeHostProblemExpire POST /hosts/<name>/cmd/acknowledge_host_problem_expire Sends the ACKNOWLEDGE_HOST_PROBLEM_EXPIRE command.
	EndpointAcknowledgeHostProblemExpire
	// EndpointAddHostComment POST /hosts/<name>/cmd/add_host_comment Sends the ADD_HOST_COMMENT command.
	EndpointAddHostComment
	// EndpointChangeCustomHostVar POST /hosts/<name>/cmd/change_custom_host_var Changes the value of a custom host variable.
	EndpointChangeCustomHostVar
	// EndpointChangeHostCheckTimeperiod POST /hosts/<name>/cmd/change_host_check_timeperiod Changes the valid check period for the specified host.
	EndpointChangeHostCheckTimeperiod
	// EndpointChangeHostModattr POST /hosts/<name>/cmd/change_host_modattr Sends the CHANGE_HOST_MODATTR command.
	EndpointChangeHostModattr
	// EndpointChangeHostNotificationTimeperiod POST /hosts/<name>/cmd/change_host_notification_timeperiod Changes the host notification timeperiod to what is specified by the 'notification_timeperiod' option. The 'notification_timeperiod' option should be the short name of the timeperiod that is to be used as the service notification timeperiod. The timeperiod must have been configured in Naemon before it was last (re)started.
	EndpointChangeHostNotificationTimeperiod
	// EndpointChangeMaxHostCheckAttempts POST /hosts/<name>/cmd/change_max_host_check_attempts Changes the maximum number of check attempts (retries) for a particular host.
	EndpointChangeMaxHostCheckAttempts
	// EndpointChangeNormalHostCheckInterval POST /hosts/<name>/cmd/change_normal_host_check_interval Changes the normal (regularly scheduled) check interval for a particular host.
	EndpointChangeNormalHostCheckInterval
	// EndpointChangeRetryHostCheckInterval POST /hosts/<name>/cmd/change_retry_host_check_interval Changes the retry check interval for a particular host.
	EndpointChangeRetryHostCheckInterval
	// EndpointDeleteActiveHostDowntimes POST /hosts/<name>/cmd/del_active_host_downtimes Removes all currently active downtimes for this host.
	EndpointDeleteActiveHostDowntimes
	// EndpointDeleteAllHostComments POST /hosts/<name>/cmd/del_all_host_comments Sends the DEL_ALL_HOST_COMMENTS command.
	EndpointDeleteAllHostComments
	// EndpointDeleteComment POST /hosts/<name>/cmd/del_comment Removes downtime by id for this host.
	EndpointDeleteComment
	// EndpointDeleteDowntime POST /hosts/<name>/cmd/del_downtime Removes downtime by id for this host.
	EndpointDeleteDowntime
	// EndpointDelayHostNotification POST /hosts/<name>/cmd/delay_host_notification Sends the DELAY_HOST_NOTIFICATION command.
	EndpointDelayHostNotification
	// EndpointDisableAllNotificationsBeyondHost POST /hosts/<name>/cmd/disable_all_notifications_beyond_host Sends the DISABLE_ALL_NOTIFICATIONS_BEYOND_HOST command.
	EndpointDisableAllNotificationsBeyondHost
	// EndpointDisableHostAndChildNotifications POST /hosts/<name>/cmd/disable_host_and_child_notifications Sends the DISABLE_HOST_AND_CHILD_NOTIFICATIONS command.
	EndpointDisableHostAndChildNotifications
	// EndpointDisableHostCheck POST /hosts/<name>/cmd/disable_host_check Sends the DISABLE_HOST_CHECK command.
	EndpointDisableHostCheck
	// EndpointDisableHostEventHandler POST /hosts/<name>/cmd/disable_host_event_handler Sends the DISABLE_HOST_EVENT_HANDLER command.
	EndpointDisableHostEventHandler
	// EndpointDisableHostFlapDetection POST /hosts/<name>/cmd/disable_host_flap_detection Sends the DISABLE_HOST_FLAP_DETECTION command.
	EndpointDisableHostFlapDetection
	// EndpointDisableHostNotifications POST /hosts/<name>/cmd/disable_host_notifications Sends the DISABLE_HOST_NOTIFICATIONS command.
	EndpointDisableHostNotifications
	// EndpointDisableHostSVCChecks POST /hosts/<name>/cmd/disable_host_svc_checks Sends the DISABLE_HOST_SVC_CHECKS command.
	EndpointDisableHostSVCChecks
	// EndpointDisableHostSVCNotifications POST /hosts/<name>/cmd/disable_host_svc_notifications Sends the DISABLE_HOST_SVC_NOTIFICATIONS command.
	EndpointDisableHostSVCNotifications
	// EndpointDisablePassiveHostChecks POST /hosts/<name>/cmd/disable_passive_host_checks Sends the DISABLE_PASSIVE_HOST_CHECKS command.
	EndpointDisablePassiveHostChecks
	// EndpointEnableAllNotificationsBeyondHost POST /hosts/<name>/cmd/enable_all_notifications_beyond_host Sends the ENABLE_ALL_NOTIFICATIONS_BEYOND_HOST command.
	EndpointEnableAllNotificationsBeyondHost
	// EndpointEnableHostAndChildNotifications POST /hosts/<name>/cmd/enable_host_and_child_notifications Sends the ENABLE_HOST_AND_CHILD_NOTIFICATIONS command.
	EndpointEnableHostAndChildNotifications
	// EndpointEnableHostCheck POST /hosts/<name>/cmd/enable_host_check Sends the ENABLE_HOST_CHECK command.
	EndpointEnableHostCheck
	// EndpointEnableHostEventHandler POST /hosts/<name>/cmd/enable_host_event_handler Sends the ENABLE_HOST_EVENT_HANDLER command.
	EndpointEnableHostEventHandler
	// EndpointEnableHostFlapDetection POST /hosts/<name>/cmd/enable_host_flap_detection Sends the ENABLE_HOST_FLAP_DETECTION command.
	EndpointEnableHostFlapDetection
	// EndpointEnableHostNotifications POST /hosts/<name>/cmd/enable_host_notifications Sends the ENABLE_HOST_NOTIFICATIONS command.
	EndpointEnableHostNotifications
	// EndpointEnableHostSVCChecks POST /hosts/<name>/cmd/enable_host_svc_checks Sends the ENABLE_HOST_SVC_CHECKS command.
	EndpointEnableHostSVCChecks
	// EndpointEnableHostSVCNotifications POST /hosts/<name>/cmd/enable_host_svc_notifications Sends the ENABLE_HOST_SVC_NOTIFICATIONS command.
	EndpointEnableHostSVCNotifications
	// EndpointEnablePassiveHostChecks POST /hosts/<name>/cmd/enable_passive_host_checks Sends the ENABLE_PASSIVE_HOST_CHECKS command.
	EndpointEnablePassiveHostChecks
	// EndpointAddHostNote POST /hosts/<name>/cmd/note Add host note to core log.
	EndpointAddHostNote
	// EndpointProcessHostCheckResult POST /hosts/<name>/cmd/process_host_check_result Sends the PROCESS_HOST_CHECK_RESULT command.
	EndpointProcessHostCheckResult
	// EndpointRemoveHostAcknowledgement POST /hosts/<name>/cmd/remove_host_acknowledgement Sends the REMOVE_HOST_ACKNOWLEDGEMENT command.
	EndpointRemoveHostAcknowledgement
	// EndpointScheduleAndPropagateHostDowntime POST /hosts/<name>/cmd/schedule_and_propagate_host_downtime Sends the SCHEDULE_AND_PROPAGATE_HOST_DOWNTIME command.
	EndpointScheduleAndPropagateHostDowntime
	// EndpointScheduleAndPropagateTriggeredHostDowntime POST /hosts/<name>/cmd/schedule_and_propagate_triggered_host_downtime Sends the SCHEDULE_AND_PROPAGATE_TRIGGERED_HOST_DOWNTIME command.
	EndpointScheduleAndPropagateTriggeredHostDowntime
	// EndpointScheduleForcedHostCheck POST /hosts/<name>/cmd/schedule_forced_host_check Sends the SCHEDULE_FORCED_HOST_CHECK command.
	EndpointScheduleForcedHostCheck
	// EndpointScheduleForcedHostSVCChecks POST /hosts/<name>/cmd/schedule_forced_host_svc_checks Sends the SCHEDULE_FORCED_HOST_SVC_CHECKS command.
	EndpointScheduleForcedHostSVCChecks
	// EndpointScheduleHostCheck POST /hosts/<name>/cmd/schedule_host_check Sends the SCHEDULE_HOST_CHECK command.
	EndpointScheduleHostCheck
	// EndpointScheduleHostDowntime POST /hosts/<name>/cmd/schedule_host_downtime Sends the SCHEDULE_HOST_DOWNTIME command.
	EndpointScheduleHostDowntime
	// EndpointScheduleHostSVCChecks POST /hosts/<name>/cmd/schedule_host_svc_checks Sends the SCHEDULE_HOST_SVC_CHECKS command.
	EndpointScheduleHostSVCChecks
	// EndpointScheduleHostSVCdowntime POST /hosts/<name>/cmd/schedule_host_svc_downtime Sends the SCHEDULE_HOST_SVC_DOWNTIME command.
	EndpointScheduleHostSVCdowntime
	// EndpointSendCustomHostNotification POST /hosts/<name>/cmd/send_custom_host_notification Sends the SEND_CUSTOM_HOST_NOTIFICATION command.
	EndpointSendCustomHostNotification
	// EndpointSetHostNotificationNumber POST /hosts/<name>/cmd/set_host_notification_number Sets the current notification number for a particular host. A value of 0 indicates that no notification has yet been sent for the current host problem. Useful for forcing an escalation (based on notification number) or replicating notification information in redundant monitoring environments. Notification numbers greater than zero have no noticeable affect on the notification process if the host is currently in an UP state.
	EndpointSetHostNotificationNumber
	// EndpointStartObsessingOverHost POST /hosts/<name>/cmd/start_obsessing_over_host Sends the START_OBSESSING_OVER_HOST command.
	EndpointStartObsessingOverHost
	// EndpointStopObsessingOverHost POST /hosts/<name>/cmd/stop_obsessing_over_host Sends the STOP_OBSESSING_OVER_HOST command.
	EndpointStopObsessingOverHost
	// EndpointGetHostCommandline GET /hosts/<name>/commandline displays commandline for check command of given hosts.
	EndpointGetHostCommandline
	// EndpointGetHostConfig GET /hosts/<name>/config Returns configuration for given host.
	EndpointGetHostConfig
	// EndpointReplaceHostConfig POST /hosts/<name>/config Replace host configuration completely, use PATCH to only update specific attributes.
	EndpointReplaceHostConfig
	// EndpointPatchHostConfig PATCH /hosts/<name>/config Update host configuration partially.
	EndpointPatchHostConfig
	// EndpointDeleteHostConfig DELETE /hosts/<name>/config Deletes given host from configuration.
	EndpointDeleteHostConfig
	// EndpointListHostNotifications GET /hosts/<name>/notifications lists notifications for given host.
	EndpointListHostNotifications
	// EndpointGetHostOutages GET /hosts/<name>/outages list of outages for this host.
	EndpointGetHostOutages
	// EndpointListHostServices GET /hosts/<name>/services lists services for given host.
	EndpointListHostServices
	// EndpointGetHostsAvailability GET /hosts/availability list availability for all hosts.
	EndpointGetHostsAvailability
	// EndpointGetHostsOutages GET /hosts/outages list of outages for all hosts.
	EndpointGetHostsOutages
	// EndpointGetHostStats GET /hosts/stats hash of livestatus host statistics.
	EndpointGetHostStats
	// EndpointGetHostTotals GET /hosts/totals hash of livestatus host totals statistics.
	EndpointGetHostTotals
	// EndpointListIndex GET /index lists all available rest urls.
	EndpointListIndex
	// EndpointListLMDSites GET /lmd/sites lists connected sites. Only available if LMD (`use_lmd`) is enabled.
	EndpointListLMDSites
	// EndpointListLogs GET /logs lists livestatus logs.
	EndpointListLogs
	// EndpointListNotifications GET /notifications lists notifications based on logfiles.
	EndpointListNotifications
	// EndpointListProcessInfo GET /processinfo lists livestatus sites status.
	EndpointListProcessInfo
	// EndpointListProcessInfoStats GET /processinfo/stats lists livestatus sites statistics.
	EndpointListProcessInfoStats
	// EndpointListServiceGroups GET /servicegroups lists livestatus servicegroups.
	EndpointListServiceGroups
	// EndpointListServiceGroupByName GET /servicegroups/<name> lists servicegroups for given name.
	EndpointListServiceGroupByName
	// EndpointGetServiceGroupAvailability GET /servicegroups/<name>/availability list availability for this servicegroup.
	EndpointGetServiceGroupAvailability
	// EndpointDisableServiceGroupHostChecks POST /servicegroups/<name>/cmd/disable_servicegroup_host_checks Sends the DISABLE_SERVICEGROUP_HOST_CHECKS command.
	EndpointDisableServiceGroupHostChecks
	// EndpointDisableServiceGroupHostNotifications POST /servicegroups/<name>/cmd/disable_servicegroup_host_notifications Sends the DISABLE_SERVICEGROUP_HOST_NOTIFICATIONS command.
	EndpointDisableServiceGroupHostNotifications
	// EndpointDisableServiceGroupPassiveHostChecks POST /servicegroups/<name>/cmd/disable_servicegroup_passive_host_checks Disables the acceptance and processing of passive checks for all hosts that have services that are members of a particular service group.
	EndpointDisableServiceGroupPassiveHostChecks
	// EndpointDisableServiceGroupPassiveSVCChecks POST /servicegroups/<name>/cmd/disable_servicegroup_passive_svc_checks Disables the acceptance and processing of passive checks for all services in a particular servicegroup.
	EndpointDisableServiceGroupPassiveSVCChecks
	// EndpointDisableServiceGroupSVCChecks POST /servicegroups/<name>/cmd/disable_servicegroup_svc_checks Sends the DISABLE_SERVICEGROUP_SVC_CHECKS command.
	EndpointDisableServiceGroupSVCChecks
	// EndpointDisableServiceGroupSVCNotifications POST /servicegroups/<name>/cmd/disable_servicegroup_svc_notifications Sends the DISABLE_SERVICEGROUP_SVC_NOTIFICATIONS command.
	EndpointDisableServiceGroupSVCNotifications
	// EndpointEnableServiceGroupHostChecks POST /servicegroups/<name>/cmd/enable_servicegroup_host_checks Sends the ENABLE_SERVICEGROUP_HOST_CHECKS command.
	EndpointEnableServiceGroupHostChecks
	// EndpointEnableServiceGroupHostNotifications POST /servicegroups/<name>/cmd/enable_servicegroup_host_notifications Sends the ENABLE_SERVICEGROUP_HOST_NOTIFICATIONS command.
	EndpointEnableServiceGroupHostNotifications
	// EndpointEnableServiceGroupPassiveHostChecks POST /servicegroups/<name>/cmd/enable_servicegroup_passive_host_checks Enables the acceptance and processing of passive checks for all hosts that have services that are members of a particular service group.
	EndpointEnableServiceGroupPassiveHostChecks
	// EndpointEnableServiceGroupPassiveSVCChecks POST /servicegroups/<name>/cmd/enable_servicegroup_passive_svc_checks Enables the acceptance and processing of passive checks for all services in a particular servicegroup.
	EndpointEnableServiceGroupPassiveSVCChecks
	// EndpointEnableServiceGroupSVCChecks POST /servicegroups/<name>/cmd/enable_servicegroup_svc_checks Sends the ENABLE_SERVICEGROUP_SVC_CHECKS command.
	EndpointEnableServiceGroupSVCChecks
	// EndpointEnableServiceGroupSVCNotifications POST /servicegroups/<name>/cmd/enable_servicegroup_svc_notifications Sends the ENABLE_SERVICEGROUP_SVC_NOTIFICATIONS command.
	EndpointEnableServiceGroupSVCNotifications
	// EndpointScheduleServiceGroupHostDowntime POST /servicegroups/<name>/cmd/schedule_servicegroup_host_downtime Sends the SCHEDULE_SERVICEGROUP_HOST_DOWNTIME command.
	EndpointScheduleServiceGroupHostDowntime
	// EndpointScheduleServiceGroupSVCdowntime POST /servicegroups/<name>/cmd/schedule_servicegroup_svc_downtime Sends the SCHEDULE_SERVICEGROUP_SVC_DOWNTIME command.
	EndpointScheduleServiceGroupSVCdowntime
	// EndpointGetServiceGroupConfig GET /servicegroups/<name>/config Returns configuration for given servicegroup.
	EndpointGetServiceGroupConfig
	// EndpointReplaceServiceGroupConfig POST /servicegroups/<name>/config Replace servicegroup configuration completely, use PATCH to only update specific attributes.
	EndpointReplaceServiceGroupConfig
	// EndpointPatchServiceGroupConfig PATCH /servicegroups/<name>/config Update servicegroup configuration partially.
	EndpointPatchServiceGroupConfig
	// EndpointDeleteServiceGroupConfig DELETE /servicegroups/<name>/config Deletes given servicegroup from configuration.
	EndpointDeleteServiceGroupConfig
	// EndpointGetServiceGroupOutages GET /servicegroups/<name>/outages list of outages for this servicegroup.
	EndpointGetServiceGroupOutages
	// EndpointGetServiceGroupStats GET /servicegroups/<name>/stats hash of livestatus servicegroup statistics.
	EndpointGetServiceGroupStats
	// EndpointListServices GET /services lists livestatus services.
	EndpointListServices
	// EndpointListServiceByHostAndName GET /services/<host>/<service> lists services for given host and name.
	EndpointListServiceByHostAndName
	// EndpointGetServiceAvailability GET /services/<host>/<service>/availability list of outages for this service.
	// EndpointGetServiceAvailability Note: description says "availability", but URL path indicates "outages" based on context. Kept as-is per spec.
	EndpointGetServiceAvailability
	// EndpointAcknowledgeSVCProblem POST /services/<host>/<service>/cmd/acknowledge_svc_problem Sends the ACKNOWLEDGE_SVC_PROBLEM command.
	EndpointAcknowledgeSVCProblem
	// EndpointAcknowledgeSVCProblemExpire POST /services/<host>/<service>/cmd/acknowledge_svc_problem_expire Sends the ACKNOWLEDGE_SVC_PROBLEM_EXPIRE command.
	EndpointAcknowledgeSVCProblemExpire
	// EndpointAddSVCComment POST /services/<host>/<service>/cmd/add_svc_comment Sends the ADD_SVC_COMMENT command.
	EndpointAddSVCComment
	// EndpointChangeCustomSVCVar POST /services/<host>/<service>/cmd/change_custom_svc_var Changes the value of a custom service variable.
	EndpointChangeCustomSVCVar
	// EndpointChangeMaxSVCCheckAttempts POST /services/<host>/<service>/cmd/change_max_svc_check_attempts Changes the maximum number of check attempts (retries) for a particular service.
	EndpointChangeMaxSVCCheckAttempts
	// EndpointChangeNormalSVCCheckInterval POST /services/<host>/<service>/cmd/change_normal_svc_check_interval Changes the normal (regularly scheduled) check interval for a particular service
	EndpointChangeNormalSVCCheckInterval
	// EndpointChangeRetrySVCCheckInterval POST /services/<host>/<service>/cmd/change_retry_svc_check_interval Changes the retry check interval for a particular service.
	EndpointChangeRetrySVCCheckInterval
	// EndpointChangeSVCCheckTimeperiod POST /services/<host>/<service>/cmd/change_svc_check_timeperiod Changes the check timeperiod for a particular service to what is specified by the 'check_timeperiod' option. The 'check_timeperiod' option should be the short name of the timeperod that is to be used as the service check timeperiod. The timeperiod must have been configured in Naemon before it was last (re)started.
	EndpointChangeSVCCheckTimeperiod
	// EndpointChangeSVCModattr POST /services/<host>/<service>/cmd/change_svc_modattr Sends the CHANGE_SVC_MODATTR command.
	EndpointChangeSVCModattr
	// EndpointChangeSVCNotificationTimeperiod POST /services/<host>/<service>/cmd/change_svc_notification_timeperiod Changes the service notification timeperiod to what is specified by the 'notification_timeperiod' option. The 'notification_timeperiod' option should be the short name of the timeperiod that is to be used as the service notification timeperiod. The timeperiod must have been configured in Naemon before it was last (re)started.
	EndpointChangeSVCNotificationTimeperiod
	// EndpointDeleteActiveServiceDowntimes POST /services/<host>/<service>/cmd/del_active_service_downtimes Removes all currently active downtimes for this service.
	EndpointDeleteActiveServiceDowntimes
	// EndpointDeleteAllSVCComments POST /services/<host>/<service>/cmd/del_all_svc_comments Sends the DEL_ALL_SVC_COMMENTS command.
	EndpointDeleteAllSVCComments
	// EndpointDeleteSVCComment POST /services/<host>/<service>/cmd/del_comment Removes downtime by id for this service.
	EndpointDeleteSVCComment
	// EndpointDeleteSVCDowntime POST /services/<host>/<service>/cmd/del_downtime Removes downtime by id for this service.
	EndpointDeleteSVCDowntime
	// EndpointDelaySVCNotification POST /services/<host>/<service>/cmd/delay_svc_notification Sends the DELAY_SVC_NOTIFICATION command.
	EndpointDelaySVCNotification
	// EndpointDisablePassiveSVCChecks POST /services/<host>/<service>/cmd/disable_passive_svc_checks Sends the DISABLE_PASSIVE_SVC_CHECKS command.
	EndpointDisablePassiveSVCChecks
	// EndpointDisableSVCCheck POST /services/<host>/<service>/cmd/disable_svc_check Sends the DISABLE_SVC_CHECK command.
	EndpointDisableSVCCheck
	// EndpointDisableSVCEventHandler POST /services/<host>/<service>/cmd/disable_svc_event_handler Sends the DISABLE_SVC_EVENT_HANDLER command.
	EndpointDisableSVCEventHandler
	// EndpointDisableSVCFlapDetection POST /services/<host>/<service>/cmd/disable_svc_flap_detection Sends the DISABLE_SVC_FLAP_DETECTION command.
	EndpointDisableSVCFlapDetection
	// EndpointDisableSVCNotifications POST /services/<host>/<service>/cmd/disable_svc_notifications Sends the DISABLE_SVC_NOTIFICATIONS command.
	EndpointDisableSVCNotifications
	// EndpointEnablePassiveSVCChecks POST /services/<host>/<service>/cmd/enable_passive_svc_checks Sends the ENABLE_PASSIVE_SVC_CHECKS command.
	EndpointEnablePassiveSVCChecks
	// EndpointEnableSVCCheck POST /services/<host>/<service>/cmd/enable_svc_check Sends the ENABLE_SVC_CHECK command.
	EndpointEnableSVCCheck
	// EndpointEnableSVCEventHandler POST /services/<host>/<service>/cmd/enable_svc_event_handler Sends the ENABLE_SVC_EVENT_HANDLER command.
	EndpointEnableSVCEventHandler
	// EndpointEnableSVCFlapDetection POST /services/<host>/<service>/cmd/enable_svc_flap_detection Sends the ENABLE_SVC_FLAP_DETECTION command.
	EndpointEnableSVCFlapDetection
	// EndpointEnableSVCNotifications POST /services/<host>/<service>/cmd/enable_svc_notifications Sends the ENABLE_SVC_NOTIFICATIONS command.
	EndpointEnableSVCNotifications
	// EndpointAddSVCNote POST /services/<host>/<service>/cmd/note Add service note to core log.
	EndpointAddSVCNote
	// EndpointProcessServiceCheckResult POST /services/<host>/<service>/cmd/process_service_check_result Sends the PROCESS_SERVICE_CHECK_RESULT command.
	EndpointProcessServiceCheckResult
	// EndpointRemoveSVCAcknowledgement POST /services/<host>/<service>/cmd/remove_svc_acknowledgement Sends the REMOVE_SVC_ACKNOWLEDGEMENT command.
	EndpointRemoveSVCAcknowledgement
	// EndpointScheduleForcedSVCCheck POST /services/<host>/<service>/cmd/schedule_forced_svc_check Sends the SCHEDULE_FORCED_SVC_CHECK command.
	EndpointScheduleForcedSVCCheck
	// EndpointScheduleSVCCheck POST /services/<host>/<service>/cmd/schedule_svc_check Sends the SCHEDULE_SVC_CHECK command.
	EndpointScheduleSVCCheck
	// EndpointScheduleSVCDowntime POST /services/<host>/<service>/cmd/schedule_svc_downtime Sends the SCHEDULE_SVC_DOWNTIME command.
	EndpointScheduleSVCDowntime
	// EndpointSendCustomSVCNotification POST /services/<host>/<service>/cmd/send_custom_svc_notification Sends the SEND_CUSTOM_SVC_NOTIFICATION command.
	EndpointSendCustomSVCNotification
	// EndpointSetSVCNotificationNumber POST /services/<host>/<service>/cmd/set_svc_notification_number Sets the current notification number for a particular service. A value of 0 indicates that no notification has yet been sent for the current service problem. Useful for forcing an escalation (based on notification number) or replicating notification information in redundant monitoring environments. Notification numbers greater than zero have no noticeable affect on the notification process if the service is currently in an OK state.
	EndpointSetSVCNotificationNumber
	// EndpointStartObsessingOverSVC POST /services/<host>/<service>/cmd/start_obsessing_over_svc Sends the START_OBSESSING_OVER_SVC command.
	EndpointStartObsessingOverSVC
	// EndpointStopObsessingOverSVC POST /services/<host>/<service>/cmd/stop_obsessing_over_svc Sends the STOP_OBSESSING_OVER_SVC command.
	EndpointStopObsessingOverSVC
	// EndpointGetServiceCommandline GET /services/<host>/<service>/commandline displays commandline for check command of given services.
	EndpointGetServiceCommandline
	// EndpointGetServiceConfig GET /services/<host>/<service>/config Returns configuration for given service.
	EndpointGetServiceConfig
	// EndpointReplaceServiceConfig POST /services/<host>/<service>/config Replace service configuration completely, use PATCH to only update specific attributes.
	EndpointReplaceServiceConfig
	// EndpointPatchServiceConfig PATCH /services/<host>/<service>/config Update service configuration partially.
	EndpointPatchServiceConfig
	// EndpointDeleteServiceConfig DELETE /services/<host>/<service>/config Deletes given service from configuration.
	EndpointDeleteServiceConfig
	// EndpointGetServiceOutages GET /services/<host>/<service>/outages list of outages for this service.
	EndpointGetServiceOutages
	// EndpointGetServicesAvailability GET /services/availability list availability for all services.
	EndpointGetServicesAvailability
	// EndpointGetServicesOutages GET /services/outages list of outages for all services.
	EndpointGetServicesOutages
	// EndpointGetServiceStats GET /services/stats livestatus service statistics.
	EndpointGetServiceStats
	// EndpointGetServiceTotals GET /services/totals livestatus service totals statistics.
	EndpointGetServiceTotals
	// EndpointListSites GET /sites lists configured backends
	EndpointListSites
	// EndpointChangeGlobalHostEventHandler POST /system/cmd/change_global_host_event_handler Changes the global host event handler command to be that specified by the 'event_handler_command' option. The 'event_handler_command' option specifies the short name of the command that should be used as the new host event handler. The command must have been configured in Naemon before it was last (re)started.
	EndpointChangeGlobalHostEventHandler
	// EndpointChangeGlobalSVCEventHandler POST /system/cmd/change_global_svc_event_handler Changes the global service event handler command to be that specified by the 'event_handler_command' option. The 'event_handler_command' option specifies the short name of the command that should be used as the new service event handler. The command must have been configured in Naemon before it was last (re)started.
	EndpointChangeGlobalSVCEventHandler
	// EndpointDeleteDowntimeByHostName POST /system/cmd/del_downtime_by_host_name This command deletes all downtimes matching the specified filters.
	EndpointDeleteDowntimeByHostName
	// EndpointDeleteDowntimeByHostgroupName POST /system/cmd/del_downtime_by_hostgroup_name This command deletes all downtimes matching the specified filters.
	EndpointDeleteDowntimeByHostgroupName
	// EndpointDeleteDowntimeByStartTimeComment POST /system/cmd/del_downtime_by_start_time_comment This command deletes all downtimes matching the specified filters.
	EndpointDeleteDowntimeByStartTimeComment
	// EndpointDelHostComment POST /system/cmd/del_host_comment Sends the DEL_HOST_COMMENT command.
	EndpointDelHostComment
	// EndpointDelHostDowntime POST /system/cmd/del_host_downtime Sends the DEL_HOST_DOWNTIME command.
	EndpointDelHostDowntime
	// EndpointDelSVCComment POST /system/cmd/del_svc_comment Sends the DEL_SVC_COMMENT command.
	EndpointDelSVCComment
	// EndpointDelSVCDowntime POST /system/cmd/del_svc_downtime Sends the DEL_SVC_DOWNTIME command.
	EndpointDelSVCDowntime
	// EndpointDisableEventHandlers POST /system/cmd/disable_event_handlers Sends the DISABLE_EVENT_HANDLERS command.
	EndpointDisableEventHandlers
	// EndpointDisableFlapDetection POST /system/cmd/disable_flap_detection Sends the DISABLE_FLAP_DETECTION command.
	EndpointDisableFlapDetection
	// EndpointDisableHostFreshnessChecks POST /system/cmd/disable_host_freshness_checks Disables freshness checks of all hosts on a program-wide basis.
	EndpointDisableHostFreshnessChecks
	// EndpointDisableNotifications POST /system/cmd/disable_notifications Sends the DISABLE_NOTIFICATIONS command.
	EndpointDisableNotifications
	// EndpointDisablePerformanceData POST /system/cmd/disable_performance_data Sends the DISABLE_PERFORMANCE_DATA command.
	EndpointDisablePerformanceData
	// EndpointDisableServiceFreshnessChecks POST /system/cmd/disable_service_freshness_checks Disables freshness checks of all services on a program-wide basis.
	EndpointDisableServiceFreshnessChecks
	// EndpointEnableEventHandlers POST /system/cmd/enable_event_handlers Sends the ENABLE_EVENT_HANDLERS command.
	EndpointEnableEventHandlers
	// EndpointEnableFlapDetection POST /system/cmd/enable_flap_detection Sends the ENABLE_FLAP_DETECTION command.
	EndpointEnableFlapDetection
	// EndpointEnableHostFreshnessChecks POST /system/cmd/enable_host_freshness_checks Enables freshness checks of all services on a program-wide basis. Individual services that have freshness checks disabled will not be checked for freshness.
	EndpointEnableHostFreshnessChecks
	// EndpointEnableNotifications POST /system/cmd/enable_notifications Sends the ENABLE_NOTIFICATIONS command.
	EndpointEnableNotifications
	// EndpointEnablePerformanceData POST /system/cmd/enable_performance_data Sends the ENABLE_PERFORMANCE_DATA command.
	EndpointEnablePerformanceData
	// EndpointEnableServiceFreshnessChecks POST /system/cmd/enable_service_freshness_checks Enables freshness checks of all services on a program-wide basis. Individual services that have freshness checks disabled will not be checked for freshness.
	EndpointEnableServiceFreshnessChecks
	// EndpointAddCustomLogEntry POST /system/cmd/log Add custom log entry to core log.
	EndpointAddCustomLogEntry
	// EndpointReadStateInformation POST /system/cmd/read_state_information Causes Naemon to load all current monitoring status information from the state retention file. Normally, state retention information is loaded when the Naemon process starts up and before it starts monitoring. WARNING: This command will cause Naemon to discard all current monitoring status information and use the information stored in state retention file! Use with care.
	EndpointReadStateInformation
	// EndpointRestartProcess POST /system/cmd/restart_process Sends the RESTART_PROCESS command.
	EndpointRestartProcess
	// EndpointRestartProgram POST /system/cmd/restart_program Restarts the Naemon process.
	EndpointRestartProgram
	// EndpointSaveStateInformation POST /system/cmd/save_state_information Causes Naemon to save all current monitoring status information to the state retention file. Normally, state retention
	EndpointSaveStateInformation
	// EndpointShutdownProcess POST /system/cmd/shutdown_process Sends the SHUTDOWN_PROCESS command.
	EndpointShutdownProcess
	// EndpointShutdownProgram POST /system/cmd/shutdown_program Shuts down the Naemon process.
	EndpointShutdownProgram
	// EndpointStartAcceptingPassiveHostChecks POST /system/cmd/start_accepting_passive_host_checks Sends the START_ACCEPTING_PASSIVE_HOST_CHECKS command.
	EndpointStartAcceptingPassiveHostChecks
	// EndpointStartAcceptingPassiveSVCChecks POST /system/cmd/start_accepting_passive_svc_checks Sends the START_ACCEPTING_PASSIVE_SVC_CHECKS command.
	EndpointStartAcceptingPassiveSVCChecks
	// EndpointStartExecutingHostChecks POST /system/cmd/start_executing_host_checks Sends the START_EXECUTING_HOST_CHECKS command.
	EndpointStartExecutingHostChecks
	// EndpointStartExecutingSVCChecks POST /system/cmd/start_executing_svc_checks Sends the START_EXECUTING_SVC_CHECKS command.
	EndpointStartExecutingSVCChecks
	// EndpointStartObsessingOverHostChecks POST /system/cmd/start_obsessing_over_host_checks Sends the START_OBSESSING_OVER_HOST_CHECKS command.
	EndpointStartObsessingOverHostChecks
	// EndpointStartObsessingOverSVCChecks POST /system/cmd/start_obsessing_over_svc_checks Sends the START_OBSESSING_OVER_SVC_CHECKS command.
	EndpointStartObsessingOverSVCChecks
	// EndpointStopAcceptingPassiveHostChecks POST /system/cmd/stop_accepting_passive_host_checks Sends the STOP_ACCEPTING_PASSIVE_HOST_CHECKS command.
	EndpointStopAcceptingPassiveHostChecks
	// EndpointStopAcceptingPassiveSVCChecks POST /system/cmd/stop_accepting_passive_svc_checks Sends the STOP_ACCEPTING_PASSIVE_SVC_CHECKS command.
	EndpointStopAcceptingPassiveSVCChecks
	// EndpointStopExecutingHostChecks POST /system/cmd/stop_executing_host_checks Sends the STOP_EXECUTING_HOST_CHECKS command.
	EndpointStopExecutingHostChecks
	// EndpointStopExecutingSVCChecks POST /system/cmd/stop_executing_svc_checks Sends the STOP_EXECUTING_SVC_CHECKS command.
	EndpointStopExecutingSVCChecks
	// EndpointStopObsessingOverHostChecks POST /system/cmd/stop_obsessing_over_host_checks Sends the STOP_OBSESSING_OVER_HOST_CHECKS command.
	EndpointStopObsessingOverHostChecks
	// EndpointStopObsessingOverSVCChecks POST /system/cmd/stop_obsessing_over_svc_checks Sends the STOP_OBSESSING_OVER_SVC_CHECKS command.
	EndpointStopObsessingOverSVCChecks
	// EndpointGetThrukInfo GET /thruk hash of basic information about this thruk instance
	EndpointGetThrukInfo
	// EndpointListAPIKeys GET /thruk/api_keys lists api keys
	EndpointListAPIKeys
	// EndpointCreateAPIKey POST /thruk/api_keys create new api key.
	EndpointCreateAPIKey
	// EndpointGetAPIKeyByID GET /thruk/api_keys/<id> alias for /thruk/api_keys?hashed_key=<id>
	EndpointGetAPIKeyByID
	// EndpointDeleteAPIKeyByID DELETE /thruk/api_keys/<id> remove key for given id.
	EndpointDeleteAPIKeyByID
	// EndpointListBusinessProcesses GET /thruk/bp lists business processes.
	EndpointListBusinessProcesses
	// EndpointCreateBusinessProcess POST /thruk/bp create new business process.
	EndpointCreateBusinessProcess
	// EndpointGetBusinessProcessByID GET /thruk/bp/<nr> business processes for given number.
	EndpointGetBusinessProcessByID
	// EndpointReplaceBusinessProcessConfig POST /thruk/bp/<nr> update business processes configuration for given number.
	EndpointReplaceBusinessProcessConfig
	// EndpointPatchBusinessProcessConfig PATCH /thruk/bp/<nr> update business processes configuration partially for given number.
	EndpointPatchBusinessProcessConfig
	// EndpointDeleteBusinessProcess DELETE /thruk/bp/<nr> remove business processes for given number.
	EndpointDeleteBusinessProcess
	// EndpointRefreshBusinessProcess POST /thruk/bp/<nr>/refresh recalculate business processes status for given number.
	EndpointRefreshBusinessProcess
	// EndpointListBroadcasts GET /thruk/broadcasts lists broadcasts
	EndpointListBroadcasts
	// EndpointCreateBroadcast POST /thruk/broadcasts create new broadcast.
	EndpointCreateBroadcast
	// EndpointGetBroadcastByFile GET /thruk/broadcasts/<file> alias for /thruk/broadcasts?file=<file>
	EndpointGetBroadcastByFile
	// EndpointReplaceBroadcastConfig POST /thruk/broadcasts/<file> update entire broadcast for given file.
	EndpointReplaceBroadcastConfig
	// EndpointPatchBroadcastConfig PATCH /thruk/broadcasts/<file> update attributes for given broadcast.
	EndpointPatchBroadcastConfig
	// EndpointDeleteBroadcast DELETE /thruk/broadcasts/<file> remove broadcast for given file.
	EndpointDeleteBroadcast
	// EndpointListClusterNodes GET /thruk/cluster lists cluster nodes
	EndpointListClusterNodes
	// EndpointGetClusterNodeState GET /thruk/cluster/<id> return cluster state for given node.
	EndpointGetClusterNodeState
	// EndpointGetClusterHeartbeatDeprecated GET /thruk/cluster/heartbeat should not be used, use POST method instead
	EndpointGetClusterHeartbeatDeprecated
	// EndpointSendClusterHeartbeat POST /thruk/cluster/heartbeat send cluster heartbeat to all other nodes
	EndpointSendClusterHeartbeat
	// EndpointRestartClusterNodes POST /thruk/cluster/restart restarts all cluster nodes sequentially
	EndpointRestartClusterNodes
	// EndpointGetThrukConfig GET /thruk/config lists configuration information
	EndpointGetThrukConfig
	// EndpointListThrukJobs GET /thruk/jobs lists thruk jobs.
	EndpointListThrukJobs
	// EndpointGetThrukJobStatus GET /thruk/jobs/<id> get thruk job status for given id.
	EndpointGetThrukJobStatus
	// EndpointGetThrukJobOutput GET /thruk/jobs/<id>/output get thruk job output for given id.
	EndpointGetThrukJobOutput
	// EndpointGetLogCacheStats GET /thruk/logcache/stats lists logcache statistics
	EndpointGetLogCacheStats
	// EndpointRunLogCacheDeltaUpdate POST /thruk/logcache/update runs the logcache delta update.
	EndpointRunLogCacheDeltaUpdate
	// EndpointGetThrukMetrics GET /thruk/metrics alias for /thruk/stats
	EndpointGetThrukMetrics
	// EndpointListPanoramaDashboards GET /thruk/panorama lists all panorama dashboards.
	EndpointListPanoramaDashboards
	// EndpointGetPanoramaDashboard GET /thruk/panorama/<nr> returns panorama dashboard for given number.
	EndpointGetPanoramaDashboard
	// EndpointEnablePanoramaDashboardMaintenance POST /thruk/panorama/<nr>/maintenance Puts given dashboard into maintenance mode.
	EndpointEnablePanoramaDashboardMaintenance
	// EndpointDisablePanoramaDashboardMaintenance DELETE /thruk/panorama/<nr>/maintenance removes maintenance mode from given dashboard.
	EndpointDisablePanoramaDashboardMaintenance
	// EndpointListRecurringDowntimes GET /thruk/recurring_downtimes lists recurring downtimes.
	EndpointListRecurringDowntimes
	// EndpointCreateRecurringDowntime POST /thruk/recurring_downtimes create new downtime.
	EndpointCreateRecurringDowntime
	// EndpointGetRecurringDowntimeByFile GET /thruk/recurring_downtimes/<file> alias for /thruk/recurring_downtimes?file=<file>
	EndpointGetRecurringDowntimeByFile
	// EndpointReplaceRecurringDowntimeConfig POST /thruk/recurring_downtimes/<file> update entire downtime for given file.
	EndpointReplaceRecurringDowntimeConfig
	// EndpointPatchRecurringDowntimeConfig PATCH /thruk/recurring_downtimes/<file> update attributes for given downtime.
	EndpointPatchRecurringDowntimeConfig
	// EndpointDeleteRecurringDowntime DELETE /thruk/recurring_downtimes/<file> remove downtime for given file.
	EndpointDeleteRecurringDowntime
	// EndpointListReports GET /thruk/reports list of reports.
	EndpointListReports
	// EndpointCreateReport POST /thruk/reports create new report.
	EndpointCreateReport
	// EndpointGetReport GET /thruk/reports/<nr> report for given number.
	EndpointGetReport
	// EndpointReplaceReportConfig POST /thruk/reports/<nr> update entire report for given number.
	EndpointReplaceReportConfig
	// EndpointPatchReportConfig PATCH /thruk/reports/<nr> update attributes for given number.
	EndpointPatchReportConfig
	// EndpointDeleteReport DELETE /thruk/reports/<nr> remove report for given number.
	EndpointDeleteReport
	// EndpointGenerateReport POST /thruk/reports/<nr>/generate generate report for given number.
	EndpointGenerateReport
	// EndpointGetReportFile GET /thruk/reports/<nr>/report return the actual report file in binary format.
	EndpointGetReportFile
	// EndpointListThrukSessions GET /thruk/sessions lists thruk sessions.
	EndpointListThrukSessions
	// EndpointGetThrukSessionStatus GET /thruk/sessions/<id> get thruk sessions status for given id.
	EndpointGetThrukSessionStatus
	// EndpointGetThrukStats GET /thruk/stats lists thruk statistics.
	EndpointGetThrukStats
	// EndpointListThrukUsers GET /thruk/users lists thruk user profiles.
	EndpointListThrukUsers
	// EndpointGetThrukUserProfile GET /thruk/users/<id> get thruk profile for given user.
	EndpointGetThrukUserProfile
	// EndpointLockThrukUser POST /thruk/users/<id>/cmd/lock lock given thruk user.
	EndpointLockThrukUser
	// EndpointUnlockThrukUser POST /thruk/users/<id>/cmd/unlock unlock given thruk user.
	EndpointUnlockThrukUser
	// EndpointGetMyProfile GET /thruk/whoami show current profile information.
	EndpointGetMyProfile
	// EndpointListTimeperiods GET /timeperiods lists livestatus timeperiods.
	EndpointListTimeperiods
	// EndpointListTimeperiodByName GET /timeperiods/<name> lists timeperiods for given name.
	EndpointListTimeperiodByName
	// EndpointGetTimeperiodConfig GET /timeperiods/<name>/config Returns configuration for given timeperiod.
	EndpointGetTimeperiodConfig
	// EndpointReplaceTimeperiodConfig POST /timeperiods/<name>/config Replace timeperiod configuration completely, use PATCH to only update specific attributes.
	EndpointReplaceTimeperiodConfig
	// EndpointPatchTimeperiodConfig PATCH /timeperiods/<name>/config Update timeperiods configuration partially.
	EndpointPatchTimeperiodConfig
	// EndpointDeleteTimeperiodConfig DELETE /timeperiods/<name>/config Deletes given timeperiod from configuration.
	EndpointDeleteTimeperiodConfig
)
