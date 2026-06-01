package workflow

type EventKind string

const (
	EventCommitNoChanges               EventKind = "commit_no_changes"
	EventRepositoryContext             EventKind = "repository_context"
	EventCommitAnalyzing               EventKind = "commit_analyzing"
	EventCommitMessageGenerated        EventKind = "commit_message_generated"
	EventCommitMessageUpdated          EventKind = "commit_message_updated"
	EventCommitCancelled               EventKind = "commit_cancelled"
	EventCommitSucceeded               EventKind = "commit_succeeded"
	EventCommitYoloStarted             EventKind = "commit_yolo_started"
	EventCommitYoloSucceeded           EventKind = "commit_yolo_succeeded"
	EventPullRequestMainUnavailable    EventKind = "pull_request_main_unavailable"
	EventPullRequestPreparationFailed  EventKind = "pull_request_preparation_failed"
	EventPullRequestNotGitHub          EventKind = "pull_request_not_github"
	EventPullRequestSkippedOnMain      EventKind = "pull_request_skipped_on_main"
	EventPullRequestCliUnavailable     EventKind = "pull_request_cli_unavailable"
	EventPullRequestCliUnauthenticated EventKind = "pull_request_cli_unauthenticated"
	EventPullRequestPublishing         EventKind = "pull_request_publishing"
	EventPullRequestPublishFailed      EventKind = "pull_request_publish_failed"
	EventPullRequestCreating           EventKind = "pull_request_creating"
	EventPullRequestBodyFallback       EventKind = "pull_request_body_fallback"
	EventPullRequestTitleFallback      EventKind = "pull_request_title_fallback"
	EventPullRequestCreateFailed       EventKind = "pull_request_create_failed"
	EventPullRequestCreated            EventKind = "pull_request_created"
	EventBranchNoChanges               EventKind = "branch_no_changes"
	EventBranchAnalyzing               EventKind = "branch_analyzing"
	EventBranchNameGenerated           EventKind = "branch_name_generated"
	EventBranchCreated                 EventKind = "branch_created"
	EventProviderFailed                EventKind = "provider_failed"
)

type Event struct {
	Kind    EventKind
	Text    string
	Err     error
	Context *PromptContext
}

type Output interface {
	Emit(Event)
}

type noopOutput struct{}

func (noopOutput) Emit(Event) {}

func ensureOutput(output Output) Output {
	if output != nil {
		return output
	}

	return noopOutput{}
}
