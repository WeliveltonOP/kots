import * as React from "react";
import PropTypes from "prop-types";
import { getGitOpsServiceSite } from "../../utilities/utilities";
import Loader from "../shared/Loader";

import "../../scss/components/gitops/GitOpsDeploymentManager.scss";

class GitOpsRepoDetails extends React.Component {
  static propTypes = {
    appName: PropTypes.string.isRequired,
    selectedService: PropTypes.object.isRequired,
    hostname: PropTypes.string,
    ownerRepo: PropTypes.string,
    branch: PropTypes.string,
    path: PropTypes.string,
    action: PropTypes.string,
    format: PropTypes.string,
    stepTitle: PropTypes.string,
  }

  static defaultProps = {
    hostname: "",
    ownerRepo: "",
    branch: "",
    path: "",
    action: "commit",
    format: "single"
  }

  constructor(props) {
    super(props);

    const {
      appName,
      selectedService,
      hostname = "",
      ownerRepo = "",
      branch = "",
      path = "",
      action = "commit",
      format = "single",
    } = this.props;

    this.state = {
      appName,
      selectedService,
      providerError: null,
      hostname,
      ownerRepo,
      branch,
      path,
      action,
      format,
      finishingSetup: false,
      showFinishedConfirm: false,
    };
  }

  componentDidMount() {
    this._mounted = true;
  }

  componentWillUnmount() {
    this._mounted = false;
  }

  onActionTypeChange = (e) => {
    if (e.target.classList.contains("js-preventDefault")) { return }
    this.setState({ action: e.target.value });
  }

  onFileContainChange = (e) => {
    if (e.target.classList.contains("js-preventDefault")) { return }
    this.setState({ format: e.target.value });
  }

  isValid = () => {
    const { ownerRepo, selectedService } = this.state;
    const provider = selectedService?.value;
    if (provider !== "other" && !ownerRepo.length) {
      this.setState({
        providerError: {
          field: "ownerRepo"
        }
      });
      return false;
    }
    return true;
  }

  onFinishSetup = async () => {
    if (!this.isValid() || !this.props.onFinishSetup) {
      return;
    }
    
    this.setState({ finishingSetup: true });

    const repoDetails = {
      ownerRepo: this.state.ownerRepo,
      branch: this.state.branch,
      path: this.state.path,
      action: this.state.action,
      format: this.state.format,
    };

    const success = await this.props.onFinishSetup(repoDetails);
    if (this._mounted) {
      if (success) {
        this.setState({ finishingSetup: false, showFinishedConfirm: true });
        setTimeout(() => {
          this.setState({ showFinishedConfirm: false });
        }, 3000);
      } else {
        this.setState({ finishingSetup: false });
      }
    }
  }

  allowUpdate = () => {
    const {
      ownerRepo,
      branch,
      path,
      action,
      format,
      selectedService
    } = this.state;
    const provider = selectedService?.value;
    if (provider === "other") {
      return true;
    }
    const isAllowed = ownerRepo !== this.props.ownerRepo || branch !== this.props.branch || path !== this.props.path || action !== this.props.action || format !== this.props.format;
    return isAllowed;
  }

  render() {
    const {
      appName,
      selectedService,
      providerError,
      hostname,
      ownerRepo,
      branch,
      path,
      action,
      format,
      finishingSetup,
      showFinishedConfirm,
    } = this.state;

    const provider = selectedService?.value;
    const serviceSite = getGitOpsServiceSite(provider, hostname);
    const isBitbucketServer = provider === "bitbucket_server";

    return (
      <div key={`action-active`} className="GitOpsDeploy--step u-textAlign--left">
          <div>
            <p className="step-title">{this.props.stepTitle || `Enable GitOps for ${appName}`}</p>

            <div className="flex flex1 u-marginBottom--30 u-marginTop--20">
              {provider !== "other" &&
                <div className="flex flex1 flex-column u-marginRight--20">
                  <p className="u-fontSize--large u-textColor--primary u-fontWeight--bold u-lineHeight--normal">{isBitbucketServer ? "Project & Repository" : "Owner & Repository"}</p>
                  <p className="u-fontSize--normal u-textColor--bodyCopy u-fontWeight--medium u-lineHeight--normal u-marginBottom--10">Where will the commit be made?</p>
                  <input type="text" className={`Input ${providerError?.field === "ownerRepo" && "has-error"}`} placeholder={isBitbucketServer ? "project/repository" : "owner/repository"} value={ownerRepo} onChange={(e) => this.setState({ ownerRepo: e.target.value })} autoFocus />
                  {providerError?.field === "ownerRepo" && <p className="u-fontSize--small u-marginTop--5 u-color--chestnut u-fontWeight--medium u-lineHeight--normal">{isBitbucketServer ? "A project and repository must be provided" : "An owner and repository must be provided"}</p>}
                </div>
              }
              {provider !== "other" &&
                <div className="flex flex1 flex-column u-marginRight--20">
                  <p className="u-fontSize--large u-textColor--primary u-fontWeight--bold u-lineHeight--normal">Branch</p>
                  <p className="u-fontSize--normal u-textColor--bodyCopy u-fontWeight--medium u-lineHeight--normal u-marginBottom--10">Leave blank to use the default branch.</p>
                  <input type="text" className={`Input`} placeholder="main" value={branch} onChange={(e) => this.setState({ branch: e.target.value })} />
                </div>
              }
              {provider !== "other" &&
                <div className="flex flex1 flex-column">
                  <p className="u-fontSize--large u-textColor--primary u-fontWeight--bold u-lineHeight--normal">Path</p>
                  <p className="u-fontSize--normal u-textColor--bodyCopy u-fontWeight--medium u-lineHeight--normal u-marginBottom--10">Where will the deployment file live?</p>
                  <input type="text" className={"Input"} placeholder="/my-path" value={path} onChange={(e) => this.setState({ path: e.target.value })} />
                </div>
              }
            </div>

            <p className="step-sub">When an update is available{appName ? ` to ${appName}` : ""}, how should the updates YAML be delivered to{selectedService?.label === "Other" ? " your GitOps provider" : ` ${serviceSite}`}?</p>
            <div className="flex flex1 u-marginTop--normal gitops-checkboxes justifyContent--center u-marginBottom--30">
              <div className="BoxedCheckbox-wrapper flex1 u-textAlign--left u-marginRight--10">
                <div className={`BoxedCheckbox flex-auto flex ${action === "commit" ? "is-active" : ""}`}>
                  <input
                    type="radio"
                    className="u-cursor--pointer hidden-input"
                    id="commitOption"
                    checked={action === "commit"}
                    defaultValue="commit"
                    onChange={this.onActionTypeChange}
                  />
                  <label htmlFor="commitOption" className="flex1 flex u-width--full u-position--relative u-cursor--pointer u-userSelect--none">
                    <div className="flex-auto">
                      <span className="icon clickable commitOptionIcon u-marginRight--10" />
                    </div>
                    <div className="flex1">
                      <p className="u-textColor--primary u-fontSize--normal u-fontWeight--medium">Create a commit</p>
                      <p className="u-textColor--bodyCopy u-fontSize--small u-fontWeight--medium u-marginTop--5">Automatic commits to repo</p>
                    </div>
                  </label>
                </div>
              </div>
              <div className="BoxedCheckbox-wrapper flex1" />
            </div>

            <div className="u-marginBottom--10 u-textAlign--left">
              <p className="u-fontSize--large u-textColor--primary u-fontWeight--bold u-lineHeight--normal">What content will it contain?</p>
              <p className="u-fontSize--normal u-textColor--bodyCopy u-fontWeight--medium u-lineHeight--normal u-marginBottom--10">Your commit can include a single rendered yaml file or it’s full output.</p>
            </div>

            <div className="flex flex1 u-marginTop--normal gitops-checkboxes justifyContent--center u-marginBottom--30">
              <div className="BoxedCheckbox-wrapper flex1 u-textAlign--left u-marginRight--10">
                <div className={`BoxedCheckbox flex1 flex ${format === "single" ? "is-active" : ""}`}>
                  <input
                    type="radio"
                    className="u-cursor--pointer hidden-input"
                    id="singleOption"
                    checked={format === "single"}
                    defaultValue="single"
                    onChange={this.onFileContainChange}
                  />
                  <label htmlFor="singleOption" className="flex1 flex u-width--full u-position--relative u-cursor--pointer u-userSelect--none">
                    <div className="flex-auto">
                      <span className="icon clickable singleOptionIcon u-marginRight--10" />
                    </div>
                    <div className="flex1">
                      <p className="u-textColor--primary u-fontSize--normal u-fontWeight--medium">Rendered YAML</p>
                      <p className="u-textColor--bodyCopy u-fontSize--small u-fontWeight--medium u-marginTop--5">Apply using <span className="inline-code no-bg">kubectl apply -f</span></p>
                    </div>
                  </label>
                </div>
              </div>
              <div className="BoxedCheckbox-wrapper flex1"></div>
            </div>

            <div className="flex">
            {this.props.showCancelBtn && <button className="btn secondary dustyGray u-marginRight--10" type="button" onClick={this.props.onCancel}>Cancel</button>}
              <button
                className="btn primary blue"
                type="button"
                disabled={finishingSetup || !this.allowUpdate()}
                onClick={this.onFinishSetup}
              >
                {finishingSetup ? this.props.ctaLoadingText : this.props.ctaText}
              </button>
              {finishingSetup && <Loader className="u-marginLeft--5" size="30" />}
              {showFinishedConfirm &&
                <div className="u-marginLeft--10 flex alignItems--center">
                  <span className="icon checkmark-icon" />
                  <span className="u-marginLeft--5 u-fontSize--small u-fontWeight--medium u-textColor--success">Settings saved</span>
                </div>
              }
            </div>
          </div>
      </div>
    );
  }
}

export default GitOpsRepoDetails;
