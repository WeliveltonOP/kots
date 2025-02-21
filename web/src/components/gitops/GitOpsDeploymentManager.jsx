import * as React from "react";
import Select from "react-select";
import find from "lodash/find";
import isEmpty from "lodash/isEmpty";
import classNames from "classnames";
import Loader from "../shared/Loader";
import ErrorModal from "../modals/ErrorModal";
import { withRouter, Link } from "react-router-dom";
import GitOpsFlowIllustration from "./GitOpsFlowIllustration";
import GitOpsRepoDetails from "./GitOpsRepoDetails";
import { getGitOpsUri, requiresHostname, Utilities } from "../../utilities/utilities";

import "../../scss/components/gitops/GitOpsDeploymentManager.scss";

const STEPS = [
  {
    step: "setup",
    title: "Set up GitOps",
  },
  {
    step: "provider",
    title: "GitOps provider",
  },
  {
    step: "action",
    title: "GitOps action ",
  },
];

const SERVICES = [
  {
    value: "github",
    label: "GitHub",
  },
  {
    value: "github_enterprise",
    label: "GitHub Enterprise",
  },
  {
    value: "gitlab",
    label: "GitLab",
  },
  {
    value: "gitlab_enterprise",
    label: "GitLab Enterprise",
  },
  {
    value: "bitbucket",
    label: "Bitbucket",
  },
  {
    value: "bitbucket_server",
    label: "Bitbucket Server",
  },
  // {
  //   value: "other",
  //   label: "Other",
  // }
];

const BITBUCKET_SERVER_DEFAULT_HTTP_PORT = "7990";
const BITBUCKET_SERVER_DEFAULT_SSH_PORT = "7999";

class GitOpsDeploymentManager extends React.Component {
  state = {
    step: "setup",
    hostname: "",
    httpPort: "",
    sshPort: "",
    services: SERVICES,
    selectedService: SERVICES[0],
    providerError: null,
    finishingSetup: false,
    appsList: [],
    gitops: {},
    errorMsg: "",
    errorTitle: "",
    displayErrorModal: false,
  }

  componentDidMount() {
    this.getAppsList();
    this.getGitops();
  }

  getAppsList = async () => {
    try {
      const res = await fetch(`${process.env.API_ENDPOINT}/apps`, {
        headers: {
          "Authorization": Utilities.getToken(),
          "Content-Type": "application/json",
        },
        method: "GET",
      });
      if (!res.ok) {
        if (res.status === 401) {
          Utilities.logoutUser();
          return;
        }
        console.log("failed to get apps list, unexpected status code", res.status);
        return;
      }
      const response = await res.json();
      const apps = response.apps;
      this.setState({
        appsList: apps,
      });
      return apps;
    } catch(err) {
      console.log(err);
      throw err;
    }
  }

  getGitops = async () => {
    try {
      const res = await fetch(`${process.env.API_ENDPOINT}/gitops/get`, {
        headers: {
          "Authorization": Utilities.getToken(),
          "Content-Type": "application/json",
        },
        method: "GET",
      });
      if (!res.ok) {
        if (res.status === 401) {
          Utilities.logoutUser();
          return;
        }
        console.log("failed to get gitops settings, unexpected status code", res.status);
        return;
      }

      const freshGitops = await res.json()

      if (freshGitops?.enabled) {
        const selectedService = find(SERVICES, service => service.value === freshGitops.provider);
        this.setState({
          selectedService: selectedService ? selectedService : this.state.selectedService,
          hostname: freshGitops.hostname || "",
          httpPort: freshGitops.httpPort || "",
          sshPort: freshGitops.sshPort || "",
          gitops: freshGitops
        })
      } else {
        this.setState({
          gitops: freshGitops
        })
      }
    } catch(err) {
      console.log(err);
      throw err;
    }
  }

  isSingleApp = () => {
    return this.state.appsList?.length === 1;
  }

  providerChanged = () => {
    const { selectedService } = this.state;
    return selectedService?.value !== this.state.gitops?.provider;
  }

  hostnameChanged = () => {
    const { hostname, selectedService } = this.state;
    const provider = selectedService?.value;
    const savedHostname = this.state.gitops?.hostname || "";
    return !this.providerChanged() && requiresHostname(provider) && hostname !== savedHostname;
  }

  httpPortChanged = () => {
    const { httpPort, selectedService } = this.state;
    const provider = selectedService?.value;
    const savedHttpPort = this.state.gitops?.httpPort || "";
    const isBitbucketServer = provider === "bitbucket_server";
    return !this.providerChanged() && isBitbucketServer && httpPort !== savedHttpPort;
  }

  sshPortChanged = () => {
    const { sshPort, selectedService } = this.state;
    const provider = selectedService?.value;
    const savedSshPort = this.state.gitops?.sshPort || "";
    const isBitbucketServer = provider === "bitbucket_server";
    return !this.providerChanged() && isBitbucketServer && sshPort !== savedSshPort;
  }

  getGitOpsInput = (provider, uri, branch, path, format, action, hostname, httpPort, sshPort) => {
    let gitOpsInput = new Object();
    gitOpsInput.provider = provider;
    gitOpsInput.uri = uri;
    gitOpsInput.branch = branch || "";
    gitOpsInput.path = path;
    gitOpsInput.format = format;
    gitOpsInput.action = action;

    if (requiresHostname(provider)) {
      gitOpsInput.hostname = hostname;
    }

    const isBitbucketServer = provider === "bitbucket_server";
    if (isBitbucketServer) {
      gitOpsInput.httpPort = httpPort;
      gitOpsInput.sshPort = sshPort;
    }

    return gitOpsInput;
  }

  finishSetup = async (repoDetails = {}) => {
    this.setState({ finishingSetup: true, errorTitle: "", errorMsg: "", displayErrorModal: false });

    const {
      ownerRepo = "",
      branch = "",
      path = "",
      action = "commit",
      format = "single"
    } = repoDetails;

    const {
      hostname,
      selectedService
    } = this.state;

    const httpPort = this.state.httpPort || BITBUCKET_SERVER_DEFAULT_HTTP_PORT;
    const sshPort = this.state.sshPort || BITBUCKET_SERVER_DEFAULT_SSH_PORT;

    const provider = selectedService.value;
    const repoUri = getGitOpsUri(provider, ownerRepo, hostname, httpPort);
    const gitOpsInput = this.getGitOpsInput(provider, repoUri, branch, path, format, action, hostname, httpPort, sshPort);

    try {
      if (this.state.gitops?.enabled && this.providerChanged()) {
        await this.resetGitOps();
      }
      await this.createGitOpsRepo(gitOpsInput);

      if (this.isSingleApp()) {
        const app = this.state.appsList[0];
        const downstream = app?.downstream;
        const clusterId = downstream?.cluster?.id;

        await this.updateAppGitOps(app.id, clusterId, gitOpsInput);

        this.props.history.push(`/app/${app.slug}/gitops`);
      } else {
        this.setState({ step: "", finishingSetup: false });
        this.getAppsList();
        this.getGitops();
      }

      return true;
    } catch (err) {
      console.log(err);
      this.setState({
        errorTitle: "Failed to finish gitops setup",
        errorMsg: err ? err.message : "Something went wrong, please try again.",
        displayErrorModal: true,
      });
      return false;
    } finally {
      this.setState({ finishingSetup: false });
    }
  }

  createGitOpsRepo = async (gitOpsInput) => {
    const res = await fetch(`${process.env.API_ENDPOINT}/gitops/create`, {
      headers: {
        "Authorization": Utilities.getToken(),
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        gitOpsInput: gitOpsInput,
      }),
      method: "POST",
    });
    if (!res.ok) {
      if (res.status === 401) {
        Utilities.logoutUser();
        return;
      }
      throw new Error(`Unexpected status code: ${res.status}`);
    }
  }

  resetGitOps = async () => {
    const res = await fetch(`${process.env.API_ENDPOINT}/gitops/reset`, {
      headers: {
        "Authorization": Utilities.getToken(),
        "Content-Type": "application/json",
      },
      method: "POST",
    });
    if (!res.ok) {
      if (res.status === 401) {
        Utilities.logoutUser();
        return;
      }
      throw new Error(`Unexpected status code: ${res.status}`);
    }
  }

  updateAppGitOps = async (appId, clusterId, gitOpsInput) => {
    const res = await fetch(`${process.env.API_ENDPOINT}/gitops/app/${appId}/cluster/${clusterId}/update`, {
      headers: {
        "Authorization": Utilities.getToken(),
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        gitOpsInput: gitOpsInput,
      }),
      method: "PUT",
    });
    if (!res.ok) {
      if (res.status === 401) {
        Utilities.logoutUser();
        return;
      }
      throw new Error(`Unexpected status code: ${res.status}`);
    }
  }

  updateSettings = () => {
    if (this.isSingleApp()) {
      this.stepFrom("provider", "action");
    } else {
      this.finishSetup();
    }
  }

  enableAppGitOps = async app => {
    if (!app.downstream) {
      return;
    }

    const downstream = app?.downstream;
    const gitops = downstream?.gitops;
    if (gitops?.enabled) {
      return;
    }

    if (isEmpty(this.state.gitops)) {
      return;
    }

    const { provider, hostname, httpPort, sshPort, uri } = this.state.gitops;
    const branch = "";
    const path = "";
    const format = "single";
    const action = "commit";
    const gitOpsInput = this.getGitOpsInput(provider, uri, branch, path, format, action, hostname, httpPort, sshPort);

    try {
      const clusterId = downstream?.cluster?.id;

      await this.updateAppGitOps(app.id, clusterId, gitOpsInput);
      
      this.props.history.push(`/app/${app.slug}/gitops`);
    } catch (err) {
      console.log(err);
      this.setState({
        errorTitle: "Failed to enable app gitops",
        errorMsg: err ? err.message : "Something went wrong, please try again.",
        displayErrorModal: true,
      });
    }
  }

  validStep = (step) => {
    const {
      selectedService,
      hostname,
    } = this.state;

    this.setState({ providerError: null });
    if (step === "provider") {
      const provider = selectedService.value;
      if (requiresHostname(provider) && !hostname.length) {
        this.setState({
          providerError: {
            field: "hostname"
          }
        });
        return false;
      }
    }

    return true;
  }

  stepFrom = (from, to) => {
    if (this.validStep(from)) {
      this.setState({
        step: to
      });
    }
  }

  renderIcons = (service) => {
    if (service) {
      return <span className={`icon gitopsService--${service.value}`} />;
    } else {
      return;
    }
  }

  getLabel = (service, label) => {
    return (
      <div style={{ alignItems: "center", display: "flex" }}>
        <span style={{ fontSize: 18, marginRight: "10px" }}>{this.renderIcons(service)}</span>
        <span style={{ fontSize: 14 }}>{label}</span>
      </div>
    );
  }

  handleServiceChange = (selectedService) => {
    this.setState({ selectedService });
  }

  renderGitOpsProviderSelector = (services, selectedService) => {
    return (
      <div className="flex flex1 flex-column u-marginRight--10">
        <p className="u-fontSize--large u-textColor--primary u-fontWeight--bold u-lineHeight--normal">Which GitOps provider do you use?</p>
        <p className="u-fontSize--normal u-textColor--bodyCopy u-fontWeight--medium u-lineHeight--normal u-marginBottom--10">Select the git provider you use for gitops.</p>
        <div className="u-position--relative">
          <Select
            className="replicated-select-container"
            classNamePrefix="replicated-select"
            placeholder="Select a GitOps service"
            options={services}
            isSearchable={false}
            getOptionLabel={(service) => this.getLabel(service, service.label)}
            getOptionValue={(service) => service.label}
            value={selectedService}
            onChange={this.handleServiceChange}
            isOptionSelected={(option) => { option.value === selectedService }}
          />
        </div>
      </div>
    );
  }

  renderHostName = (provider, hostname, providerError) => {
    if (!requiresHostname(provider)) {
      return <div className="flex flex1" />;
    }
    return (
      <div className="flex flex1 flex-column u-marginLeft--10">
        <p className="u-fontSize--large u-textColor--primary u-fontWeight--bold u-lineHeight--normal">Hostname</p>
        <p className="u-fontSize--normal u-textColor--bodyCopy u-fontWeight--medium u-lineHeight--normal u-marginBottom--10">Hostname of your GitOps server.</p>
        <input type="text" className={`Input ${providerError?.field === "hostname" && "has-error"}`} placeholder="hostname" value={hostname} onChange={(e) => this.setState({ hostname: e.target.value })} />
        {providerError?.field === "hostname" && <p className="u-fontSize--small u-marginTop--5 u-textColor--error u-fontWeight--medium u-lineHeight--normal">A hostname must be provided</p>}
      </div>
    );
  }

  renderHttpPort = (provider, httpPort) => {
    const isBitbucketServer = provider === "bitbucket_server";
    if (!isBitbucketServer) {
      return <div className="flex flex1" />;
    }
    return (
      <div className="flex flex1 flex-column u-marginRight--10">
        <p className="u-fontSize--large u-color--tuna u-fontWeight--bold u-lineHeight--normal">HTTP Port</p>
        <p className="u-fontSize--normal u-color--dustyGray u-fontWeight--medium u-lineHeight--normal u-marginBottom--10">HTTP Port of your GitOps server.</p>
        <input type="text" className="Input" placeholder={BITBUCKET_SERVER_DEFAULT_HTTP_PORT} value={httpPort} onChange={(e) => this.setState({ httpPort: e.target.value })} />
      </div>
    );
  }

  renderSshPort = (provider, sshPort) => {
    const isBitbucketServer = provider === "bitbucket_server";
    if (!isBitbucketServer) {
      return <div className="flex flex1" />;
    }
    return (
      <div className="flex flex1 flex-column u-marginLeft--10">
        <p className="u-fontSize--large u-color--tuna u-fontWeight--bold u-lineHeight--normal">SSH Port</p>
        <p className="u-fontSize--normal u-color--dustyGray u-fontWeight--medium u-lineHeight--normal u-marginBottom--10">SSH Port of your GitOps server.</p>
        <input type="text" className="Input" placeholder={BITBUCKET_SERVER_DEFAULT_SSH_PORT} value={sshPort} onChange={(e) => this.setState({ sshPort: e.target.value })} />
      </div>
    );
  }

  renderActiveStep = (step) => {
    const {
      hostname,
      httpPort,
      sshPort,
      services,
      selectedService,
      providerError,
      finishingSetup,
    } = this.state;

    const provider = selectedService?.value;
    const isBitbucketServer = provider === "bitbucket_server";

    switch (step.step) {
      case "setup":
        return (
        <div key={`${step.step}-active`} className="GitOpsDeploy--step">
          <p className="step-title">Deploy using a GitOps workflow</p>
          <p className="step-sub">Connect a git version control system to this Admin Console. After setting this up, it will be<br/>possible to have all application updates (upstream updates, license updates, config changes)<br/>directly commited to any git repository and it will no longer be possible to deploy directly from the Admin Console.</p>
          <GitOpsFlowIllustration />
          <div>
            <button className="btn primary blue u-marginTop--10" type="button" onClick={() => this.stepFrom("setup", "provider")}>Get started</button>
          </div>
        </div>
      );
      case "provider":
        return (
          <div key={`${step.step}-active`} className="GitOpsDeploy--step u-textAlign--left">
            <p className="step-title">{step.title}</p>
            <p className="step-sub">Before the Admin Console can push changes to your Git repository, some information about your Git configuration is required.</p>
            <div className="flex-column u-textAlign--left u-marginBottom--30">
              <div className="flex flex1">
                {this.renderGitOpsProviderSelector(services, selectedService)}
                {this.renderHostName(provider, hostname, providerError)}
              </div>
              {isBitbucketServer && 
                <div className="flex flex1 u-marginTop--30">
                  {this.renderHttpPort(provider, httpPort)}
                  {this.renderSshPort(provider, sshPort)}
                </div>
              }
            </div>
            <div>
              <button
                className="btn primary blue"
                type="button"
                disabled={finishingSetup}
                onClick={this.updateSettings}
              >
                {finishingSetup
                  ? "Finishing setup"
                  : this.isSingleApp()
                    ? "Continue to deployment action"
                    : "Finish GitOps setup"
                }
              </button>
            </div>
          </div>
        );
      case "action":
        return (
          <GitOpsRepoDetails
            appName={this.props.appName}
            hostname={hostname}
            selectedService={selectedService}
            onFinishSetup={this.finishSetup}
            ctaLoadingText="Finishing setup"
            ctaText="Finish setup"
          />
        );
      default:
        return <div key={`default-active`} className="GitOpsDeploy--step">default</div>;
    }
  }

  getGitOpsStatus = gitops => {
    if (gitops?.enabled && gitops?.isConnected) {
      return "Enabled, Working";
    }
    if (gitops?.enabled) {
      return "Enabled, Failing";
    }
    return "Not Enabled";
  }

  renderGitOpsStatusAction = (app, gitops) => {
    if (gitops?.enabled && gitops?.isConnected) {
      return null;
    }
    if (gitops?.enabled) {
      return <Link to={`/app/${app.slug}/troubleshoot`} className="gitops-action-link">Troubleshoot</Link>
    }

    return <span onClick={() => this.enableAppGitOps(app)} className="gitops-action-link">Enable</span>;
  }

  renderApps = () => {
    return (
      <div>
        {this.state.appsList.map(app => {
          const downstream = app?.downstream;
          const gitops = downstream?.gitops;
          const gitopsEnabled = gitops?.enabled;
          const gitopsConnected = gitops?.isConnected;
          return (
            <div key={app.id} className="flex justifyContent--spaceBetween alignItems--center u-marginBottom--30">
              <div className="flex alignItems--center">
                <div style={{ backgroundImage: `url(${app.iconUri})` }} className="appIcon u-position--relative" />
                <div className="u-marginLeft--10">
                  <p className="u-fontSize--large u-fontWeight--bold u-textColor--secondary u-marginBottom--5">{app.name}</p>
                  {gitopsEnabled && <Link to={`/app/${app.slug}/gitops`} className="gitops-action-link">Manage GitOps settings</Link>}
                </div>
              </div>
              <div className="flex-column alignItems--flexEnd">
                <div className="flex alignItems--center u-marginBottom--5">
                  <div className={classNames("icon", {
                    "grayCircleMinus--icon": !gitopsEnabled && !gitopsConnected,
                    "error-small": gitopsEnabled && !gitopsConnected,
                    "checkmark-icon": gitopsEnabled && gitopsConnected
                    })}
                  />
                  <p className={classNames("u-fontSize--normal u-marginLeft--5", {
                    "u-textColor--bodyCopy": !gitopsEnabled && !gitopsConnected,
                    "u-textColor--error": gitopsEnabled && !gitopsConnected,
                    "u-textColor--success": gitopsEnabled && gitopsConnected,
                  })}>
                    {this.getGitOpsStatus(gitops)}
                  </p>
                </div>
                {this.renderGitOpsStatusAction(app, gitops)}
              </div>
            </div>
          );
        })}
      </div>
    );
  }

  dataChanged = () => {
    return this.providerChanged() || this.hostnameChanged() || this.httpPortChanged() || this.sshPortChanged();
  }

  renderConfiguredGitOps = () => {
    const { services, selectedService, hostname, httpPort, sshPort, providerError, finishingSetup } = this.state;
    const provider = selectedService?.value;
    const isBitbucketServer = provider === "bitbucket_server";
    const dataChanged = this.dataChanged();
    return (
      <div className="u-textAlign--center">
        <div className="ConfiguredGitOps--wrapper">
            <p className="u-fontSize--largest u-fontWeight--bold u-textColor--accent u-lineHeight--normal u-marginBottom--30">Admin Console GitOps</p>
            <div className="flex u-marginBottom--30">
              {this.renderGitOpsProviderSelector(services, selectedService)}
              {this.renderHostName(selectedService?.value, hostname, providerError)}
            </div>
            {isBitbucketServer && 
              <div className="flex u-marginBottom--30">
                {this.renderHttpPort(selectedService?.value, httpPort)}
                {this.renderSshPort(selectedService?.value, sshPort)}
              </div>
            }
            {dataChanged &&
              <button className="btn secondary u-marginBottom--30" disabled={finishingSetup} onClick={this.updateSettings}>
                {finishingSetup ? "Updating" : "Update"}
              </button>
            }
            <div className="separator" />
            {this.renderApps()}
        </div>
      </div>
    );
  }

  toggleErrorModal = () => {
    this.setState({ displayErrorModal: !this.state.displayErrorModal });
  }

  render() {
    const { appsList, errorMsg, errorTitle, displayErrorModal } = this.state;

    if (!appsList.length) {
      return (
        <div className="flex-column flex1 alignItems--center justifyContent--center">
          <Loader size="60" />
        </div>
      );
    }

    const activeStep = find(STEPS, { step: this.state.step });
    return (
      <div className="GitOpsDeploymentManager--wrapper flex-column flex1">
        {this.state.gitops?.enabled && this.state.step !== "action" ?
          this.renderConfiguredGitOps()
          : activeStep &&
          this.renderActiveStep(activeStep)
        }

        {errorMsg &&
          <ErrorModal
            errorModal={displayErrorModal}
            toggleErrorModal={this.toggleErrorModal}
            err={errorTitle}
            errMsg={errorMsg}
          />}
      </div>
    );
  }
}

export default withRouter(GitOpsDeploymentManager);
