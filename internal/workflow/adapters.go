package workflow

import (
	"context"
	"errors"

	"github.com/madflow/kommit/internal/config"
	"github.com/madflow/kommit/internal/git"
	"github.com/madflow/kommit/internal/provider"
)

var ErrPullRequestMainBranchUnavailable = errors.New("pull request main branch is unavailable")

type GenerationOptions struct {
	Provider string
	Model    string
	Stream   bool
}

type GenerationKind string

const (
	GenerationKindCommitMessage    GenerationKind = "commit_message"
	GenerationKindBranchName       GenerationKind = "branch_name"
	GenerationKindPullRequestBody  GenerationKind = "pull_request_body"
	GenerationKindPullRequestTitle GenerationKind = "pull_request_title"
)

type GenerationRequest struct {
	Kind    GenerationKind
	Diff    string
	Rules   string
	Context *PromptContext
}

type ChangedFile struct {
	Status string
	Path   string
	Type   string
}

type PromptContext struct {
	BranchName    string
	FilesChanged  int
	ChangeSummary string
	FileChanges   []ChangedFile
}

type ChangeSet struct {
	HasChanges bool
	Diff       string
	Context    *PromptContext
}

type PullRequestSkipReason string

const (
	PullRequestSkipNone               PullRequestSkipReason = ""
	PullRequestSkipNotGitHub          PullRequestSkipReason = "not_github"
	PullRequestSkipMainBranch         PullRequestSkipReason = "main_branch"
	PullRequestSkipCliUnavailable     PullRequestSkipReason = "gh_unavailable"
	PullRequestSkipCliUnauthenticated PullRequestSkipReason = "gh_unauthenticated"
)

type PullRequestPlan struct {
	CurrentBranch string
	MainBranch    string
	CombinedDiff  string
	SkipReason    PullRequestSkipReason
}

type CommitAction string

const (
	CommitActionYes  CommitAction = "yes"
	CommitActionEdit CommitAction = "edit"
	CommitActionNo   CommitAction = "no"
)

type CommitPrompter interface {
	AskForConfirmation() (CommitAction, error)
	EditCommitMessage(message string) (string, error)
}

type CommitRepository interface {
	EnsureRepository() error
	StageAll() error
	LoadStagedSnapshot() (*ChangeSet, error)
	Commit(message string) error
	PublishCurrentBranch() error
	PreparePullRequest(localDiff string) (*PullRequestPlan, error)
	CreatePullRequest(title string, body string) (string, error)
}

type BranchRepository interface {
	EnsureRepository() error
	StageAll() error
	LoadStagedSnapshot() (*ChangeSet, error)
	CreateBranch(branchName string) error
}

type Generator interface {
	Generate(ctx context.Context, req GenerationRequest, opts GenerationOptions) (string, error)
}

type CommitDependencies struct {
	Repository CommitRepository
	Generator  Generator
	Prompter   CommitPrompter
	Output     Output
}

type BranchDependencies struct {
	Repository BranchRepository
	Generator  Generator
	Output     Output
}

type gitRepository struct{}

func (gitRepository) CreateBranch(branchName string) error {
	return git.Current().CreateBranch(branchName)
}

func (gitRepository) EnsureRepository() error {
	return git.Current().EnsureRepository()
}

func (gitRepository) StageAll() error {
	return git.Current().StageAll()
}

func (gitRepository) LoadStagedSnapshot() (*ChangeSet, error) {
	snapshot, err := git.Current().LoadStagedSnapshot()
	if err != nil {
		return nil, err
	}

	return mapChangeSet(snapshot), nil
}

func (gitRepository) Commit(message string) error {
	return git.Current().Commit(message)
}

func (gitRepository) PublishCurrentBranch() error {
	return git.Current().PublishCurrentBranch()
}

func (gitRepository) PreparePullRequest(localDiff string) (*PullRequestPlan, error) {
	plan, err := git.Current().PreparePullRequest(localDiff)
	if err != nil {
		if errors.Is(err, git.ErrOriginMainBranchUnavailable) {
			return nil, ErrPullRequestMainBranchUnavailable
		}

		return nil, err
	}

	return mapPullRequestPlan(plan), nil
}

func (gitRepository) CreatePullRequest(title string, body string) (string, error) {
	return git.Current().CreatePullRequest(title, body)
}

type providerGenerator struct {
	settings *config.ResolvedSettings
	output   Output
}

func (g providerGenerator) Generate(ctx context.Context, req GenerationRequest, opts GenerationOptions) (string, error) {
	client, generateOpts := g.clientAndOptions(opts)
	return client.Generate(ctx, provider.Request{
		Kind:          provider.Kind(req.Kind),
		Diff:          req.Diff,
		Rules:         req.Rules,
		PromptContext: mapProviderPromptContext(req.Context),
	}, generateOpts)
}

func (g providerGenerator) clientAndOptions(opts GenerationOptions) (*provider.Client, provider.GenerateOptions) {
	settings := g.settings
	providerName := settings.Provider
	if opts.Provider != "" {
		providerName = opts.Provider
	}
	modelName := settings.Model
	if opts.Model != "" {
		modelName = opts.Model
	}

	client := provider.NewClient(settings)

	var onProviderError func(string, error)
	if g.output != nil {
		if _, ok := g.output.(noopOutput); !ok {
			onProviderError = func(providerName string, err error) {
				g.output.Emit(Event{Kind: EventProviderFailed, Text: providerName, Err: err})
			}
		}
	}

	return client, provider.GenerateOptions{
		Provider:        providerName,
		Model:           modelName,
		Stream:          opts.Stream,
		OnProviderError: onProviderError,
	}
}

func ensureSettings(settings *config.ResolvedSettings) *config.ResolvedSettings {
	if settings != nil {
		return settings
	}

	return config.Get()
}

func mapChangeSet(snapshot *git.StagedSnapshot) *ChangeSet {
	if snapshot == nil {
		return &ChangeSet{Context: &PromptContext{}}
	}

	return &ChangeSet{
		HasChanges: snapshot.HasChanges,
		Diff:       snapshot.Diff,
		Context:    mapPromptContext(snapshot.Context),
	}
}

func mapPromptContext(repoCtx *git.RepoContext) *PromptContext {
	if repoCtx == nil {
		return &PromptContext{}
	}

	fileChanges := make([]ChangedFile, 0, len(repoCtx.FileChanges))
	for _, change := range repoCtx.FileChanges {
		fileChanges = append(fileChanges, ChangedFile{
			Status: change.Status,
			Path:   change.FilePath,
			Type:   change.FileType,
		})
	}

	return &PromptContext{
		BranchName:    repoCtx.BranchName,
		FilesChanged:  repoCtx.FilesChanged,
		ChangeSummary: repoCtx.ChangeSummary,
		FileChanges:   fileChanges,
	}
}

func mapPullRequestPlan(plan *git.PullRequestPreparation) *PullRequestPlan {
	if plan == nil {
		return &PullRequestPlan{}
	}

	return &PullRequestPlan{
		CurrentBranch: plan.CurrentBranch,
		MainBranch:    plan.MainBranch,
		CombinedDiff:  plan.CombinedDiff,
		SkipReason:    mapPullRequestSkipReason(plan.SkipReason),
	}
}

func mapPullRequestSkipReason(reason git.PullRequestSkipReason) PullRequestSkipReason {
	switch reason {
	case git.PullRequestSkipNotGitHub:
		return PullRequestSkipNotGitHub
	case git.PullRequestSkipMainBranch:
		return PullRequestSkipMainBranch
	case git.PullRequestSkipGhUnavailable:
		return PullRequestSkipCliUnavailable
	case git.PullRequestSkipGhUnauthenticated:
		return PullRequestSkipCliUnauthenticated
	default:
		return PullRequestSkipNone
	}
}

func mapProviderPromptContext(ctx *PromptContext) *provider.PromptContext {
	if ctx == nil {
		return &provider.PromptContext{}
	}

	fileChanges := make([]provider.ChangedFile, 0, len(ctx.FileChanges))
	for _, change := range ctx.FileChanges {
		fileChanges = append(fileChanges, provider.ChangedFile{
			Status: change.Status,
			Path:   change.Path,
			Type:   change.Type,
		})
	}

	return &provider.PromptContext{
		BranchName:   ctx.BranchName,
		FilesChanged: ctx.FilesChanged,
		Summary:      ctx.ChangeSummary,
		FileChanges:  fileChanges,
	}
}
