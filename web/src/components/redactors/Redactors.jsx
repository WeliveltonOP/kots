import React, { Component } from "react";
import { withRouter, Link } from "react-router-dom"
import Helmet from "react-helmet";
import dayjs from "dayjs";
import Select from "react-select";
import Toggle from "../shared/Toggle";
import RedactorRow from "./RedactorRow";
import DeleteRedactorModal from "../modals/DeleteRedactorModal";
import Loader from "../shared/Loader";

import { Utilities } from "../../utilities/utilities";

class Redactors extends Component {
  state = {
    redactors: [],
    sortedRedactors: [],
    selectedOption: {
      value: "enabled",
      label: "Sort by: Status"
    },
    deleteRedactorModal: false,
    redactorToDelete: {},
    isLoadingRedactors: false,
    redactorsErrMsg: "",
    deletingRedactor: false,
    deleteErrMsg: "",
    enablingRedactorMsg: ""
  };

  getRedactors = () => {
    this.setState({
      isLoadingRedactors: true,
      redactorsErrMsg: ""
    });

    fetch(`${process.env.API_ENDPOINT}/redacts`, {
      method: "GET",
      headers: {
        "Authorization": Utilities.getToken(),
        "Content-Type": "application/json",
      }
    })
      .then(res => res.json())
      .then(result => {
        if (result.success) {
          this.setState({
            redactors: result.redactors,
            isLoadingRedactors: false,
            redactorsErrMsg: "",
          }, () => {
            if (this.state.selectedOption) {
              this.sortRedactors(this.state.selectedOption.value);
            }
          })
        } else {
          this.setState({
            isLoadingRedactors: false,
            redactorsErrMsg: result.error,
          })
        }
      })
      .catch(err => {
        this.setState({
          isLoadingRedactors: false,
          redactorsErrMsg: err,
        })
      })
  }

  handleSortChange = selectedOption => {
    this.setState({ selectedOption }, () => {
      this.sortRedactors(this.state.selectedOption.value);
    });
  }

  componentDidMount() {
    this.getRedactors();
  }

  sortRedactors = value => {
    if (value === "createdAt") {
      this.setState({ sortedRedactors: this.state.redactors.sort((a, b) => dayjs(b.createdAt) - dayjs(a.createdAt)) });
    } else if (value === "updatedAt") {
      this.setState({ sortedRedactors: this.state.redactors.sort((a, b) => dayjs(b.updatedAt) - dayjs(a.updatedAt)) });
    } else {
      this.setState({ sortedRedactors: this.state.redactors.sort((a, b) => (a.enabled === b.enabled) ? 0 : a.enabled ? -1 : 1 )});
    }
  }

  toggleConfirmDeleteModal = redactor => {
    if (this.state.deleteRedactorModal) {
      this.setState({ deleteRedactorModal: false, redactorToDelete: "", deleteErr: false, deleteErrorMsg: "" });
    } else {
      this.setState({ deleteRedactorModal: true, redactorToDelete: redactor, deleteErr: false, deleteErrorMsg: "" });
    }
  };

  handleDeleteRedactor = redactor => {
    this.setState({ deletingRedactor: true, deleteErrMsg: "" });

    fetch(`${process.env.API_ENDPOINT}/redact/spec/${redactor.slug}`, {
      method: "DELETE",
      headers: {
        "Authorization": Utilities.getToken(),
        "Content-Type": "application/json"
      }
    })
      .then(() => {
        this.setState({
          deletingRedactor: false,
          deleteErrMsg: "",
          deleteRedactorModal: false
        }, () => {
          this.getRedactors();
        });
      })
      .catch((err) => {
        this.setState({
          deletingRedactor: false,
          deleteErrMsg: err.message ? err.message : "Something went wrong, please try again."
        });
      });
  }

  handleSetRedactEnabled = (redactor, redactorEnabled) => {
    const payload = {
      enabled: redactorEnabled
    }
    this.setState({ enablingRedactorMsg: "" });
    fetch(`${process.env.API_ENDPOINT}/redact/enabled/${redactor.slug}`, {
      method: "POST",
      headers: {
        "Authorization": Utilities.getToken(),
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload)
    })
      .then(async (res) => {
        const response = await res.json();
        if (!res.ok) {
          this.setState({ enablingRedactorMsg: response.error });
          return;
        }
        if (response.success) {
          this.setState({ enablingRedactorMsg: "" });
        } else {
          this.setState({ enablingRedactorMsg: response.error });
        }
      })
      .catch((err) => {
        this.setState({
          enablingRedactorMsg: err.message ? err.message : "Something went wrong, please try again."
        });
      });
  }



  render() {
    const { sortedRedactors, selectedOption, deleteRedactorModal, isLoadingRedactors, enablingRedactorMsg } = this.state;

    if (isLoadingRedactors) {
      return (
        <div className="flex-column flex1 alignItems--center justifyContent--center">
          <Loader size="60" />
        </div>
      )
    }

    const selectOptions = [
      {
        value: "enabled",
        label: "Sort by: Status"
      },
      {
        value: "createdAt",
        label: "Sort by: Created At"
      },
      {
        value: "updatedAt",
        label: "Sort by: Updated on"
      }
    ]

    return (
      <div className="container flex-column flex1 u-overflow--auto u-paddingTop--30 u-paddingBottom--20 justifyContent--center alignItems--center">
        <Helmet>
          <title>Redactors </title>
        </Helmet>
        <div className="Redactors--wrapper flex1 flex-column u-width--full">
          <div className="flex justifyContent--center u-paddingBottom--30">
            <Toggle
              items={[
                {
                  title: "Support bundles",
                  onClick: () => this.props.history.push(`/app/${this.props.appSlug}/troubleshoot`),
                  isActive: false
                },
                {
                  title: "Redactors",
                  onClick: () => this.props.history.push(`/app/${this.props.appSlug}/troubleshoot/redactors`),
                  isActive: true
                }
              ]}
            />
          </div>
          {sortedRedactors?.length > 0 ?
            <div className="flex1 flex-column">
              <div className="flex flex-auto alignItems--center justifyContent--spaceBetween">
                <div className="flex flex1 alignItems--center">
                  <p className="u-fontWeight--bold u-textColor--primary u-fontSize--larger u-lineHeight--normal u-marginRight--10">Redactors</p>
                  <div style={{ width: "220px" }}>
                    <Select
                      className="replicated-select-container"
                      classNamePrefix="replicated-select"
                      options={selectOptions}
                      value={selectedOption}
                      getOptionValue={(option) => option.label}
                      isOptionSelected={(option) => { option.value === selectedOption }}
                      onChange={this.handleSortChange}
                    />
                  </div>
                </div>
                <div className="flex justifyContent--flexEnd">
                  <Link to={`/app/${this.props.appSlug}/troubleshoot/redactors/new`} className="btn primary blue">Create new redactor</Link>
                </div>
              </div>
              <p className="u-fontSize--normal u-textColor--bodyCopy u-fontWeight--medium u-lineHeight--normal u-marginTop--20 u-marginBottom--30">Define custom rules for sensitive values you need to be redacted when gathering a support bundle. This might include things like Secrets or IP addresses. For help with creating custom redactors,
              <a href="https://troubleshoot.sh/reference/redactors/overview/" target="_blank" rel="noopener noreferrer" className="replicated-link"> check out our docs</a>.</p>
              {enablingRedactorMsg && <p className="u-textColor--error u-fontSize--normal u-fontWeight--medium u-lineHeight--normal u-marginBottom--10 flex justifyContent--center alignItems--center">{enablingRedactorMsg}</p>}
              {sortedRedactors?.map((redactor) => (
                <RedactorRow
                  key={`redactor-${redactor.slug}`}
                  redactor={redactor}
                  appSlug={this.props.appSlug}
                  toggleConfirmDeleteModal={this.toggleConfirmDeleteModal}
                  handleSetRedactEnabled={this.handleSetRedactEnabled}

                />
              ))}
            </div>
            :
            <div className="flex-column flex1 alignItems--center">
              <div className="flex-column flex1 alignItems--center u-textAlign--center justifyContent--center u-width--half">
                <p className="u-fontSize--20 u-fontWeight--bold u-textColor--secondary u-lineHeight--normal">Configure custom redactors</p>
                <p className="u-fontSize--normal u-textColor--bodyCopy u-fontWeight--medium u-lineHeight--normal u-marginTop--20">Define custom rules for sensitive values you need to be redacted when gathering a support bundle. This might include things like Secrets or IP addresses. For help with creating custom redactors,
                <a href="https://troubleshoot.sh/reference/redactors/overview/" target="_blank" rel="noopener noreferrer" className="replicated-link"> check out our docs</a>.</p>
                <div className="u-marginTop--30">
                  <Link to={`/app/${this.props.appSlug}/troubleshoot/redactors/new`} className="btn primary blue">Create new redactor</Link>
                </div>
              </div>
            </div>
          }
        </div>
        {deleteRedactorModal &&
          <DeleteRedactorModal
            deleteRedactorModal={deleteRedactorModal}
            toggleConfirmDeleteModal={this.toggleConfirmDeleteModal}
            handleDeleteRedactor={this.handleDeleteRedactor}
            redactorToDelete={this.state.redactorToDelete}
            deletingRedactor={this.state.deletingRedactor}
            deleteErrMsg={this.state.deleteErrMsg}
          />
        }
      </div>
    );
  }
}

export default withRouter(Redactors);
