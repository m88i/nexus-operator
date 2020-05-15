#     Copyright 2020 Nexus Operator and/or its authors
#
#     This file is part of Nexus Operator.
#
#     Nexus Operator is free software: you can redistribute it and/or modify
#     it under the terms of the GNU General Public License as published by
#     the Free Software Foundation, either version 3 of the License, or
#     (at your option) any later version.
#
#     Nexus Operator is distributed in the hope that it will be useful,
#     but WITHOUT ANY WARRANTY; without even the implied warranty of
#     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#     GNU General Public License for more details.
#
#     You should have received a copy of the GNU General Public License
#     along with Nexus Operator.  If not, see <https://www.gnu.org/licenses/>.

# this Dockerfile is based on https://github.com/operator-framework/community-operators/blob/master/upstream.Dockerfile
# we just changed the third line to include only our manifest in this registry. 
# since we don't have dependencies from others operators, we're good to go
FROM quay.io/operator-framework/upstream-registry-builder:v1.5.6 as builder
ARG PERMISSIVE_LOAD=true
COPY build/_output/operatorhub/ manifests
RUN if [ $PERMISSIVE_LOAD = "true" ] ; then ./bin/initializer --permissive -o ./bundles.db ; else ./bin/initializer -o ./bundles.db ; fi 

FROM scratch
COPY --from=builder /build/bundles.db /bundles.db
COPY --from=builder /build/bin/registry-server /registry-server
COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe
EXPOSE 50051
ENTRYPOINT ["/registry-server"]
CMD ["--database", "/bundles.db"]
