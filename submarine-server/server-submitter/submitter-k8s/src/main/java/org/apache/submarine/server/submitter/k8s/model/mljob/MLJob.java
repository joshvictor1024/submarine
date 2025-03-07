/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package org.apache.submarine.server.submitter.k8s.model.mljob;

import com.google.gson.JsonSyntaxException;
import com.google.gson.annotations.SerializedName;
import io.kubernetes.client.common.KubernetesObject;
import io.kubernetes.client.openapi.models.V1JobStatus;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import io.kubernetes.client.openapi.models.V1ObjectMetaBuilder;
import io.kubernetes.client.openapi.models.V1Status;
import org.apache.submarine.commons.utils.exception.SubmarineRuntimeException;
import org.apache.submarine.server.api.common.CustomResourceType;
import org.apache.submarine.server.api.experiment.Experiment;
import org.apache.submarine.server.api.spec.ExperimentMeta;
import org.apache.submarine.server.api.spec.ExperimentSpec;
import org.apache.submarine.server.k8s.utils.K8sUtils;
import org.apache.submarine.server.submitter.k8s.client.K8sClient;
import org.apache.submarine.server.submitter.k8s.model.K8sResource;
import org.apache.submarine.server.submitter.k8s.util.JsonUtils;
import org.apache.submarine.server.submitter.k8s.util.MLJobConverter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * The machine learning job for the CRD job.
 * It is serialized as body input to k8s api client.
 * <p>
 * For job resource definitions and related information,
 * please refer to https://github.com/kubeflow/training-operator
 */
public abstract class MLJob implements KubernetesObject, K8sResource<Experiment> {

  protected final Logger LOG = LoggerFactory.getLogger(getClass());

  public MLJob(ExperimentSpec experimentSpec) {
    // set metadata
    V1ObjectMetaBuilder metaBuilder = new V1ObjectMetaBuilder();
    metaBuilder.withNamespace(K8sUtils.getNamespace())
        // SUBMARINE-880 replace name to experimentId
        .withName(experimentSpec.getMeta().getExperimentId())
        //.addAllToOwnerReferences(OwnerReferenceUtils.getOwnerReference())
        .addToLabels(ExperimentMeta.SUBMARINE_EXPERIMENT_NAME, experimentSpec.getMeta().getName());
    setMetadata(metaBuilder.build());
    // set framework
    setFramework(experimentSpec.getMeta().getFramework());
    // set experimentId
    setExperimentId(experimentSpec.getMeta().getExperimentId());
  }

  @SerializedName("apiVersion")
  private String apiVersion;

  @SerializedName("kind")
  private String kind;

  @SerializedName("metadata")
  private V1ObjectMeta metadata;

  // transient to avoid being serialized
  private transient String group;

  private transient String version;

  private transient String plural;

  @SerializedName("status")
  private V1JobStatus status;

  private String framework;

  private String experimentId;

  /**
   * Set the api with version
   *
   * @param apiVersion api with version
   */
  public void setApiVersion(String apiVersion) {
    this.apiVersion = apiVersion;
  }

  /**
   * Get the kind
   *
   * @return kind
   */
  public String getKind() {
    return kind;
  }

  /**
   * Set the CRD's name
   *
   * @param kind the CRD's name
   */
  public void setKind(String kind) {
    this.kind = kind;
  }

  /**
   * Get the metadata
   *
   * @return meta
   */
  public V1ObjectMeta getMetadata() {
    return metadata;
  }

  /**
   * Set metadata
   *
   * @param metadata meta
   */
  public void setMetadata(V1ObjectMeta metadata) {
    this.metadata = metadata;
  }

  /**
   * Get the api with version
   *
   * @return api with version
   */
  public String getApiVersion() {
    return apiVersion;
  }

  /**
   * Get the resource's group name
   *
   * @return group name
   */
  public String getGroup() {
    return this.group;
  }

  /**
   * Get the resource's version
   *
   * @return version
   */
  public String getVersion() {
    return this.version;
  }

  /**
   * Get the resource's plural
   *
   * @return plural
   */
  public String getPlural() {
    return this.plural;
  }

  public void setGroup(String group) {
    this.group = group;
  }

  public void setVersion(String version) {
    this.version = version;
  }

  public void setPlural(String plural) {
    this.plural = plural;
  }

  public V1JobStatus getStatus() {
    return status;
  }

  public void setStatus(V1JobStatus status) {
    this.status = status;
  }

  public String getFramework() {
    return framework;
  }

  public void setFramework(String framework) {
    this.framework = framework;
  }

  public String getExperimentId() {
    return experimentId;
  }

  public void setExperimentId(String experimentId) {
    this.experimentId = experimentId;
  }

  /**
   * Convert MLJob object to return Experiment object
   */
  protected <T extends MLJob> Experiment parseExperimentResponseObject(T object, Class<T> tClass)
          throws SubmarineRuntimeException {
    String jsonString = JsonUtils.toJson(object);
    LOG.info("Upstream response JSON: {}", jsonString);
    try {
      return MLJobConverter.toJobFromMLJob(JsonUtils.fromJson(jsonString, tClass));
    } catch (JsonSyntaxException e) {
      LOG.error("K8s submitter: parse response object failed by " + e.getMessage(), e);
      throw new SubmarineRuntimeException(500, "K8s Submitter parse upstream response failed.");
    }
  }

  /**
   * Convert MLJob status to return Experiment object
   */
  protected Experiment parseExperimentResponseStatus(V1Status status)
          throws SubmarineRuntimeException {
    String jsonString = JsonUtils.toJson(status);
    LOG.info("Upstream response JSON: {}", jsonString);
    try {
      return MLJobConverter.toJobFromStatus(JsonUtils.fromJson(jsonString, V1Status.class));
    } catch (JsonSyntaxException e) {
      LOG.error("K8s submitter: parse response object failed by " + e.getMessage(), e);
      throw new SubmarineRuntimeException(500, "K8s Submitter parse upstream response failed.");
    }
  }

  /**
   * Get custom resource type
   * This method is mainly used by pod agent
   */
  public abstract CustomResourceType getResourceType();

  @Override
  public Experiment read(K8sClient api) {
    throw new UnsupportedOperationException("MLJob does not implement this method!");
  }

  @Override
  public Experiment create(K8sClient api) {
    throw new UnsupportedOperationException("MLJob does not implement this method!");
  }

  @Override
  public Experiment replace(K8sClient api) {
    throw new UnsupportedOperationException("MLJob does not implement this method!");
  }

  @Override
  public Experiment delete(K8sClient api) {
    throw new UnsupportedOperationException("MLJob does not implement this method!");
  }
}
