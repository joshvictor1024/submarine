# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements. See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License. You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
"""
TensorFlow implementation of FM

Reference:
[1] Factorization machines for CTR Prediction,
    Steffen Rendle
"""

import logging

import tensorflow as tf

from submarine.ml.tensorflow_v2.layers.core import embedding_layer, fm_layer, linear_layer
from submarine.ml.tensorflow_v2.model.base_tf_model import BaseTFModel
from submarine.utils.tf_utils_v2 import get_estimator_spec

logger = logging.getLogger(__name__)


class FM(BaseTFModel):
    def model_fn(self, features, labels, mode, params):
        super().model_fn(features, labels, mode, params)

        linear_logit = linear_layer(features, **params["training"])
        embedding_outputs = embedding_layer(features, **params["training"])
        fm_logit = fm_layer(embedding_outputs, **params["training"])

        with tf.compat.v1.variable_scope("FM_out"):
            logit = linear_logit + fm_logit

        return get_estimator_spec(logit, labels, mode, params)
