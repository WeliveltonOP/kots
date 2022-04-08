import * as React from "react";
import { withRouter, Link } from "react-router-dom";
import { Helmet } from "react-helmet";
import size from "lodash/size";
import isEmpty from "lodash/isEmpty";
import Modal from "react-modal";
import CodeSnippet from "./shared/CodeSnippet";

import "../scss/components/HelmLogin.scss";
import { Utilities } from "../utilities/utilities";

class HelmLogin extends React.Component {
  state = {
    errorMessage: "",
    selectedAppToInstall: {}
  }

  componentDidUpdate(prevProps) {
    if (this.props.isHelmManaged !== prevProps.isHelmManaged && !this.props.isHelmManaged) {
      this.props.history.replace("/upload-license");
    }
  }

  componentDidMount() {
    const { appSlugFromMetadata } = this.props;

    if (appSlugFromMetadata) {
      const hasChannelAsPartOfASlug = appSlugFromMetadata.includes("/");
      let appSlug;
      if (hasChannelAsPartOfASlug) {
        const splitAppSlug = appSlugFromMetadata.split("/");
        appSlug = splitAppSlug[0]
      } else {
        appSlug = appSlugFromMetadata;
      }
      this.setState({
        selectedAppToInstall: {
          ...this.state.selectedAppToInstall,
          value: appSlug,
          label: appSlugFromMetadata
        }
      })
    }
  }

  onLogin = async () => {
    try {
      const res = await fetch(`${process.env.API_ENDPOINT}/helm/login`, {
        headers: {
          "Authorization": Utilities.getToken(),
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          hostname: this.state.hostname,
          username: this.state.username,
          password: this.state.password,
        }),
        method: "POST",
      });

      if (res.status === 401) {
        Utilities.logoutUser();
        return;
      }
      if (res.status === 400) {
        const body = await res.json();
        console.log(body);
        return;
      }

    } catch (err) {
      console.log(err);
      throw(err);
    }
  }

  render() {
    const {
      appName,
      logo,
      fetchingMetadata,
      appsListLength,
      appSlugFromMetadata,
      isHelmManaged,
    } = this.props;
    const { licenseFile, fileUploading, errorMessage, viewErrorMessage, licenseExistErrData, selectedAppToInstall, hasMultiApp } = this.state;
    const hasFile = licenseFile && !isEmpty(licenseFile);

    let logoUri;
    let applicationName;
    if (appsListLength && appsListLength > 1) {
      logoUri = "https://cdn2.iconfinder.com/data/icons/mixd/512/16_kubernetes-512.png";
      applicationName = "";
    } else {
      logoUri = logo;
      applicationName = appSlugFromMetadata ? appSlugFromMetadata : appName;
    }

    // TODO remove when restore is enabled
    const isRestoreEnabled = false;

    if (isHelmManaged === null) {
      return null;
    }

    return (
      <div className={`UploadLicenseFile--wrapper container flex-column flex1 u-overflow--auto Login-wrapper justifyContent--center alignItems--center`}>
        <Helmet>
          <title>{`${applicationName ? `${applicationName} Admin Console` : "Admin Console"}`}</title>
        </Helmet>
        <div className="LoginBox-wrapper u-flexTabletReflow  u-flexTabletReflow flex-auto">
          <div className="flex-auto flex-column login-form-wrapper secure-console justifyContent--center">
            <div className="flex-column alignItems--center">
              {logo
                ? <span className="icon brand-login-icon" style={{ backgroundImage: `url(${logoUri})` }} />
                : !fetchingMetadata ? <span className="icon kots-login-icon" />
                  : <span style={{ width: "60px", height: "60px" }} />
              }
            </div>
              <div className="flex flex-column">
                <p className="u-fontSize--header u-textColor--primary u-fontWeight--bold u-textAlign--center u-marginTop--10 u-paddingTop--5">Log in to the Helm registry</p>
                <div className="u-marginTop--30">
                  <div className={`FileUpload-wrapper flex1 ${hasFile ? "has-file" : ""}`}>
                    <p className="u-fontSize--large u-textColor--primary u-fontWeight--bold u-lineHeight--normal">Endpoint</p>
                    <p className="u-fontSize--normal u-textColor--bodyCopy u-fontWeight--medium u-lineHeight--normal u-marginBottom--10">The hostname of the Helm registry</p>
                    <input type="text" className="Input" placeholder="registry.replicated.com" onChange={(e) => this.setState({ ownerRepo: e.target.value })} autoFocus />

                    <p className="u-fontSize--large u-textColor--primary u-fontWeight--bold u-lineHeight--normal">Username</p>
                    <p className="u-fontSize--normal u-textColor--bodyCopy u-fontWeight--medium u-lineHeight--normal u-marginBottom--10">The email address to log in with</p>
                    <input type="text" className="Input" placeholder="registry.replicated.com" onChange={(e) => this.setState({ ownerRepo: e.target.value })} autoFocus />

                    <p className="u-fontSize--large u-textColor--primary u-fontWeight--bold u-lineHeight--normal">Password</p>
                    <p className="u-fontSize--normal u-textColor--bodyCopy u-fontWeight--medium u-lineHeight--normal u-marginBottom--10">The password to log in with</p>
                    <input type="password" className="Input" placeholder="registry.replicated.com" onChange={(e) => this.setState({ ownerRepo: e.target.value })} autoFocus />

                    <button
                      className="btn primary blue"
                      type="button"
                      onClick={this.onLogin}
                    >Log in
                    </button>
                  </div>
                </div>
                {errorMessage && (
                  <div className="u-marginTop--10">
                    <span className="u-fontSize--small u-textColor--error u-marginRight--5 u-fontWeight--bold">Unable to install license</span>
                    <span
                      className="u-fontSize--small replicated-link"
                      onClick={this.toggleViewErrorMessage}>
                      view more
                    </span>
                  </div>
                )}
              </div>
          </div>
        </div>

        <Modal
          isOpen={viewErrorMessage}
          onRequestClose={this.toggleViewErrorMessage}
          contentLabel="Online install error message"
          ariaHideApp={false}
          className="Modal"
        >
          <div className="Modal-body">
            <div className="ExpandedError--wrapper u-marginTop--10 u-marginBottom--10">
              <p className="u-fontSize--small u-fontWeight--bold u-textColor--primary u-marginBottom--5">Error description</p>
              <p className="u-fontSize--small u-textColor--error">{typeof errorMessage === "object" ? "An unknown error orrcured while trying to upload your license. Please try again." : errorMessage}</p>
              {!size(licenseExistErrData) ?
                <div className="flex flex-column">
                  <p className="u-fontSize--small u-fontWeight--bold u-marginTop--15 u-textColor--primary">Run this command to generate a support bundle</p>
                  <CodeSnippet
                    language="bash"
                    canCopy={true}
                    onCopyText={<span className="u-textColor--success">Command has been copied to your clipboard</span>}
                  >
                    kubectl support-bundle https://kots.io
              </CodeSnippet>
                </div> :
                <div className="flex flex-column">
                  <p className="u-fontSize--small u-fontWeight--bold u-marginTop--15 u-textColor--primary">Run this command to remove the app</p>
                  <CodeSnippet
                    language="bash"
                    canCopy={true}
                    onCopyText={<span className="u-textColor--success">Command has been copied to your clipboard</span>}
                  >
                    {licenseExistErrData?.deleteAppCommand}
                  </CodeSnippet>
                </div>
              }
            </div>
            <button type="button" className="btn primary u-marginTop--15" onClick={this.toggleViewErrorMessage}>Ok, got it!</button>
          </div>
        </Modal>
      </div >
    );
  }
}

export default withRouter(HelmLogin);
