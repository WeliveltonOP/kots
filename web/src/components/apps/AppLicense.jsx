import React, { Component } from "react";
import Helmet from "react-helmet";
import Dropzone from "react-dropzone";
import yaml from "js-yaml";
import classNames from "classnames";
import size from "lodash/size";
import Modal from "react-modal";
import { Link } from "react-router-dom";
import { getFileContent, Utilities, getLicenseExpiryDate } from "../../utilities/utilities";
import Loader from "../shared/Loader";

import "@src/scss/components/apps/AppLicense.scss";

class AppLicense extends Component {

  constructor(props) {
    super(props);

    this.state = {
      appLicense: null,
      loading: false,
      message: "",
      messageType: "info",
      showNextStepModal: false,
      entitlementsToShow: [],
      showLicenseChangeModal: false,
      licenseChangeFile: null,
      changingLicense: false,
      licenseChangeMessage: "",
      licenseChangeMessageType: "info",
    }
  }

  getAppLicense = async () => {
    await fetch(`${process.env.API_ENDPOINT}/app/${this.props.app.slug}/license`, {
      method: "GET",
      headers: {
        "Authorization": Utilities.getToken(),
        "Content-Type": "application/json",
      }
    }).then(async (res) => {
      const body = await res.json();
      if (body === null) {
        this.setState({ appLicense: {} });
      } else if (body.success) {
        this.setState({ appLicense: body.license });
      } else if (body.error) {
        console.log(body.error);
      }
    }).catch((err) => {
      console.log(err)
    });
  }

  componentDidMount() {
    this.getAppLicense();
  }

  onDrop = async (files) => {
    const content = await getFileContent(files[0]);
    const contentStr = (new TextDecoder("utf-8")).decode(content);
    const airgapLicense = await yaml.safeLoad(contentStr);
    const { appLicense } = this.state;

    if (airgapLicense.spec?.licenseID !== appLicense?.id) {
      this.setState({
        message: "Licenses do not match",
        messageType: "error"
      });
      return;
    }

    if (airgapLicense.spec?.licenseSequence === appLicense?.licenseSequence) {
      this.setState({
        message: "License is already up to date",
        messageType: "info"
      });
      return;
    }

    this.syncAppLicense(contentStr);
  }

  syncAppLicense = (licenseData) => {
    this.setState({
      loading: true,
      message: "",
      messageType: "info",
    });

    const { app } = this.props;

    const payload = {
      licenseData,
    };

    fetch(`${process.env.API_ENDPOINT}/app/${app.slug}/license`, {
      method: "PUT",
      headers: {
        "Authorization": Utilities.getToken(),
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload)
    })
      .then(async response => {
        if (!response.ok) {
          if (response.status == 401) {
            Utilities.logoutUser();
            return;
          }
          const res = await response.json();
          throw new Error(res?.error);
        }
        return response.json();
      })
      .then(async (licenseResponse) => {
        let message;
        if (!licenseResponse.synced) {
          message = "License is already up to date"
        } else if (app.isAirgap) {
          message = "License uploaded successfully"
        } else {
          message = "License synced successfully"
        }

        this.setState({
          appLicense: licenseResponse.license,
          message,
          messageType: "info",
          showNextStepModal: licenseResponse.synced
        });

        if (this.props.syncCallback) {
          this.props.syncCallback();
        }
      })
      .catch(err => {
        console.log(err);
        this.setState({
          message: err ? err.message : "Something went wrong",
          messageType: "error"
        });
      })
      .finally(() => {
        this.setState({ loading: false });
      });
  }

  onLicenseChangeDrop = async (files) => {
    this.setState({
      licenseChangeFile: files[0],
      licenseChangeMessage: "",
    });
  }

  clearLicenseChangeFile = () => {
    this.setState({ licenseChangeFile: null, licenseChangeMessage: "" });
  }

  changeAppLicense = async () => {
    if (!this.state.licenseChangeFile) {
      return;
    }

    const content = await getFileContent(this.state.licenseChangeFile);
    const licenseData = (new TextDecoder("utf-8")).decode(content);

    this.setState({
      changingLicense: true,
      licenseChangeMessage: "",
      licenseChangeMessageType: "info",
    });

    const { app } = this.props;

    const payload = {
      licenseData,
    };

    fetch(`${process.env.API_ENDPOINT}/app/${app.slug}/change-license`, {
      method: "PUT",
      headers: {
        "Authorization": Utilities.getToken(),
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload)
    })
      .then(async response => {
        if (!response.ok) {
          if (response.status == 401) {
            Utilities.logoutUser();
            return;
          }
          const res = await response.json();
          throw new Error(res?.error);
        }
        return response.json();
      })
      .then(async (licenseResponse) => {
        this.setState({
          appLicense: licenseResponse.license,
          showNextStepModal: true,
          showLicenseChangeModal: false,
          licenseChangeFile: null,
          licenseChangeMessage: "",
        });

        if (this.props.changeCallback) {
          this.props.changeCallback();
        }
      })
      .catch(err => {
        console.log(err);
        this.setState({
          licenseChangeMessage: err ? err.message : "Something went wrong",
          licenseChangeMessageType: "error"
        });
      })
      .finally(() => {
        this.setState({ changingLicense: false });
      });
  }

  hideNextStepModal = () => {
    this.setState({ showNextStepModal: false });
  }

  hideLicenseChangeModal = () => {
    this.setState({
      showLicenseChangeModal: false,
      licenseChangeFile: null,
      licenseChangeMessage: "",
    });
  }

  showLicenseChangeModal = () => {
    this.setState({ showLicenseChangeModal: true });
  }

  toggleShowDetails = (entitlement) => {
    this.setState({ entitlementsToShow: [...this.state.entitlementsToShow, entitlement] })
  }

  toggleHideDetails = (entitlement) => {
    let entitlementsToShow = [...this.state.entitlementsToShow];
    const index = this.state.entitlementsToShow.indexOf(entitlement);
    entitlementsToShow.splice(index, 1);
    this.setState({ entitlementsToShow })
  }

  render() {
    const {
      appLicense,
      loading,
      message,
      messageType,
      showNextStepModal,
      showLicenseChangeModal,
      licenseChangeFile,
      changingLicense,
      licenseChangeMessage,
      licenseChangeMessageType
    } = this.state;

    if (!appLicense) {
      return (
        <div className="flex-column flex1 alignItems--center justifyContent--center">
          <Loader size="60" />
        </div>
      );
    }

    const { app } = this.props;
    const expiresAt = getLicenseExpiryDate(appLicense);
    const gitops = app.downstream?.gitops;
    const appName = app?.name || "Your application";


    return (
      <div className="flex flex-column justifyContent--center alignItems--center">
        <Helmet>
          <title>{`${appName} License`}</title>
        </Helmet>
        {size(appLicense) > 0 ?
          <div className="License--wrapper flex-column">
            <div className="flex flex-auto alignItems--center">
              <span className="u-fontSize--large u-fontWeight--bold u-lineHeight--normal u-textColor--primary"> License </span>
              {appLicense?.licenseType === "community" &&
                <div className="flex-auto">
                  <span className="CommunityEditionTag u-marginLeft--10"> Community Edition </span>
                  <span className="u-fontSize--small u-fontWeight--normal u-lineHeight--normal u-marginLeft--10" style={{ color: "#A5A5A5" }}> To change your license, please contact your account representative. </span>
                </div>}
            </div>
            <div className="LicenseDetails flex flex1 justifyContent--spaceBetween">
              <div className="Details--wrapper flex-auto flex-column">
                <div className="flex flex-auto alignItems--center">
                  <span className="u-fontSize--larger u-fontWeight--bold u-lineHeight--normal u-textColor--secondary"> {appLicense.assignee} </span>
                  {appLicense?.channelName &&
                    <span className="channelTag flex-auto alignItems--center u-fontWeight--medium u-marginLeft--10"> {appLicense.channelName} </span>}
                </div>
                <div className="flex flex1 alignItems--center u-marginTop--5">
                  <div className={`LicenseTypeTag ${appLicense?.licenseType} flex-auto flex-verticalCenter alignItems--center`}>
                    <span className={`icon ${appLicense?.licenseType === "---" ? "" : appLicense?.licenseType}-icon`}></span>
                    {appLicense?.licenseType !== "---"
                      ? `${Utilities.toTitleCase(appLicense.licenseType)} license`
                      : `---`}
                  </div>
                  <p className={`u-fontWeight--medium u-fontSize--small u-lineHeight--normal u-marginLeft--10 ${Utilities.checkIsDateExpired(expiresAt) ? "u-textColor--error" : "u-textColor--info"}`}>
                    {expiresAt === "Never" ? "Does not expire" : Utilities.checkIsDateExpired(expiresAt) ? `Expired ${expiresAt}` : `Expires ${expiresAt}`}
                  </p>
                </div>
                {size(appLicense?.entitlements) > 0 &&
                  <div className="flexWrap--wrap flex-auto flex1 u-marginTop--12">
                    {appLicense.entitlements?.map((entitlement, i) => {
                      const currEntitlement = this.state.entitlementsToShow.find(f => f === entitlement.title);
                      const isTextField = entitlement.valueType === "Text";
                      const isBooleanField = entitlement.valueType === "Boolean";
                      if (entitlement.value.length > 30 && (currEntitlement !== entitlement.title)) {
                        return (
                          <span key={entitlement.label} className={`u-fontSize--small u-lineHeight--normal u-textColor--secondary u-fontWeight--medium u-marginRight--10 ${i !== 0 ? "u-marginLeft--5" : ""}`}> {entitlement.title}: <span className={`u-fontWeight--bold ${isTextField && "u-fontFamily--monospace"}`}> {entitlement.value.slice(0, 30) + "..."} </span>
                            <span className="icon clickable down-arrow-icon" onClick={() => this.toggleShowDetails(entitlement.title)} />
                          </span>
                        )
                      } else if (entitlement.value.length > 30 && (currEntitlement === entitlement.title)) {
                        return (
                          <span key={entitlement.label} className={`flex-column u-fontSize--small u-lineHeight--normal u-textColor--secondary u-fontWeight--medium u-marginRight--10 ${i !== 0 ? "u-marginLeft--5" : ""}`}> {entitlement.title}: <span className={`u-fontWeight--bold ${isTextField && "u-fontFamily--monospace"}`} style={{whiteSpace: "pre"}}> {entitlement.value} </span>
                            <span className="icon clickable up-arrow-icon u-marginTop--5 u-marginLeft--5" onClick={() => this.toggleHideDetails(entitlement.title)} />
                          </span>
                        )
                      } else {
                        return (
                          <span key={entitlement.label} className={`u-fontSize--small u-lineHeight--normal u-textColor--secondary u-fontWeight--medium u-marginRight--10 ${i !== 0 ? "u-marginLeft--5" : ""}`}> {entitlement.title}: <span className={`u-fontWeight--bold ${isTextField && "u-fontFamily--monospace"}`}> {isBooleanField ? entitlement.value.toString() : entitlement.value} </span></span>
                        );
                      }
                    })}
                  </div>}
                <div className="flexWrap--wrap flex alignItems--center">
                  {appLicense?.isAirgapSupported ? <span className="flex alignItems--center u-fontWeight--medium u-fontSize--small u-lineHeight--normal u-textColor--secondary u-marginRight--10 u-marginTop--10"><span className="icon licenseAirgapIcon" /> Airgap enabled </span> : null}
                  {appLicense?.isSnapshotSupported ? <span className="flex alignItems--center u-fontWeight--medium u-fontSize--small u-lineHeight--normal u-textColor--secondary u-marginLeft--5 u-marginRight--10 u-marginTop--10"><span className="icon licenseVeleroIcon" /> Snapshots enabled </span> : null}
                  {appLicense?.isGitOpsSupported ? <span className="flex alignItems--center u-fontWeight--medium u-fontSize--small u-lineHeight--normal u-textColor--secondary u-marginLeft--5 u-marginRight--10 u-marginTop--10"><span className="icon licenseGithubIcon" /> GitOps enabled </span> : null}
                  {appLicense?.isIdentityServiceSupported ? <span className="flex alignItems--center u-fontWeight--medium u-fontSize--small u-lineHeight--normal u-textColor--secondary u-marginLeft--5 u-marginRight--10 u-marginTop--10"><span className="icon licenseIdentityIcon" /> Identity Service enabled </span> : null}
                  {appLicense?.isGeoaxisSupported ? <span className="flex alignItems--center u-fontWeight--medium u-fontSize--small u-lineHeight--normal u-textColor--secondary u-marginLeft--5 u-marginRight--10 u-marginTop--10"><span className="icon licenseGeoaxisIcon" /> GEOAxIS Provider enabled </span> : null}
                </div>
              </div>
              <div className="flex-column flex-auto alignItems--flexEnd justifyContent--center">
                <div className="flex alignItems--center">
                  {appLicense?.licenseType === "community" &&
                    <button className="btn secondary blue u-marginRight--10" disabled={changingLicense} onClick={this.showLicenseChangeModal}>{changingLicense ? "Changing" : "Change license"}</button>
                  }
                  {app.isAirgap ?
                    <Dropzone
                      className="Dropzone-wrapper"
                      accept={["application/x-yaml", ".yaml", ".yml"]}
                      onDropAccepted={this.onDrop}
                      multiple={false}
                    >
                      <button className="btn primary blue" disabled={loading}>{loading ? "Uploading" : "Upload license"}</button>
                    </Dropzone>
                    :
                    <button className="btn primary blue" disabled={loading} onClick={() => this.syncAppLicense("")}>{loading ? "Syncing" : "Sync license"}</button>
                  }
                </div>
                {message &&
                  <p className={classNames("u-fontWeight--bold u-fontSize--small u-marginTop--10", {
                    "u-textColor--error": messageType === "error",
                    "u-textColor--primary": messageType === "info",
                  })}>{message}</p>
                }
                {appLicense?.lastSyncedAt &&
                  <p className="u-fontWeight--bold u-fontSize--small u-textColor--header u-lineHeight--default u-marginTop--10">
                    Last synced {Utilities.dateFromNow(appLicense.lastSyncedAt)}
                  </p>
                }
              </div>
            </div>
          </div>
          :
          <div>
            <p className="u-fontSize--large u-textColor--bodyCopy u-marginTop--15 u-lineHeight--more"> License data is not available on this application because it was installed via Helm </p>
          </div>
        }
        <Modal
          isOpen={showNextStepModal}
          onRequestClose={this.hideNextStepModal}
          shouldReturnFocusAfterClose={false}
          contentLabel="Next step"
          ariaHideApp={false}
          className="Modal MediumSize"
        >
          {gitops?.enabled ?
            <div className="Modal-body">
              <p className="u-fontSize--large u-textColor--primary u-lineHeight--medium u-marginBottom--20">
                The license for {appName} has been updated. A new commit has been made to the gitops repository with these changes. Please head to the <a className="link" target="_blank" href={gitops?.uri} rel="noopener noreferrer">repo</a> to see the diff.
              </p>
              <div className="flex justifyContent--flexEnd">
                <button type="button" className="btn blue primary" onClick={this.hideNextStepModal}>Ok, got it!</button>
              </div>
            </div>
            :
            <div className="Modal-body">
              <p className="u-fontSize--large u-textColor--primary u-lineHeight--medium u-marginBottom--20">
                The license for {appName} has been updated. A new version is available on the version history page with these changes.
              </p>
              <div className="flex justifyContent--flexEnd">
                <button type="button" className="btn blue secondary u-marginRight--10" onClick={this.hideNextStepModal}>Cancel</button>
                <Link to={`/app/${app?.slug}/version-history`}>
                  <button type="button" className="btn blue primary">Go to new version</button>
                </Link>
              </div>
            </div>
          }
        </Modal>

        {showLicenseChangeModal &&
          <Modal
            isOpen={showLicenseChangeModal}
            onRequestClose={this.hideLicenseChangeModal}
            shouldReturnFocusAfterClose={false}
            contentLabel="Change License"
            ariaHideApp={false}
            className="Modal SmallSize"
          >
            <div className="u-marginTop--10 u-padding--20">
              <p className="u-fontSize--larger u-fontWeight--bold u-textColor--primary u-marginBottom--10">Change your license</p>
              <p className="u-fontSize--normal u-textColor--bodyCopy u-lineHeight--normal u-marginBottom--10">The new license must be for the same application as your current license.</p>
              <div className={`FileUpload-wrapper flex1 ${licenseChangeFile ? "has-file" : ""}`}>
                {licenseChangeFile ?
                  <div className="has-file-wrapper">
                    <div className="flex">
                      <div className="icon u-yamlLtGray-small u-marginRight--10" />
                      <div>
                        <p className="u-fontSize--normal u-textColor--primary u-fontWeight--medium">{licenseChangeFile.name}</p>
                        <span className="replicated-link u-fontSize--small" onClick={this.clearLicenseChangeFile}>Select a different file</span>
                      </div>
                    </div>
                  </div>
                  :
                  <Dropzone
                    className="Dropzone-wrapper"
                    accept={["application/x-yaml", ".yaml", ".yml"]}
                    onDropAccepted={this.onLicenseChangeDrop}
                    multiple={false}
                  >
                    <div className="u-textAlign--center">
                      <div className="icon u-yamlLtGray-lrg u-marginBottom--10" />
                      <p className="u-fontSize--normal u-textColor--secondary u-fontWeight--medium u-lineHeight--normal">Drag your new license here or <span className="u-linkColor u-fontWeight--medium u-textDecoration--underlineOnHover">choose a file</span></p>
                      <p className="u-fontSize--small u-textColor--bodyCopy u-fontWeight--normal u-lineHeight--normal u-marginTop--10">This will be a .yaml file. Please contact your account rep if you are unable to locate your new license file.</p>
                    </div>
                  </Dropzone>
                }
              </div>
              {licenseChangeMessage &&
                <p className={classNames("u-fontWeight--bold u-fontSize--small u-marginTop--10 u-marginBottom--20", {
                  "u-textColor--error": licenseChangeMessageType === "error",
                  "u-textColor--primary": licenseChangeMessageType === "info",
                })}>{licenseChangeMessage}</p>
              }
              <div className="flex flex-auto">
                <button
                  type="button"
                  className="btn secondary large u-marginRight--10"
                  onClick={this.hideLicenseChangeModal}
                >
                  Cancel
                </button>
                {licenseChangeFile &&
                  <button
                    type="button"
                    className="btn primary large"
                    onClick={this.changeAppLicense}
                    disabled={changingLicense}
                  >
                    {changingLicense ? "Changing" : "Change license"}
                  </button>
                }
              </div>
            </div>
          </Modal>
        }
      </div>
    );
  }
}

export default (AppLicense);
