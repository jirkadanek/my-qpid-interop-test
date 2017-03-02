#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#

# Building qpid-proton-go shim requires Go version 1.x or later
#
# Define -DBUILD_GO=ON (or OFF) at the cmake command line
# to force a build or not.

project (qpid-interop-test-amqpnetlite-shims)

# to have ${CMAKE_COMMAND} -E env
cmake_minimum_required(VERSION 3.2.0 FATAL_ERROR)

# Set BUILD_GO control variable based on sensed environment

option(go_default ON)

#find_program(PROG_GO go)
#if (PROG_GO STREQUAL "PROG_GO-NOTFOUND")
#  message(STATUS "Program 'go' is not found. qpid-proton-go shim requires go.")
#  set(go_default OFF)
#endif ()

# Set option that controls the build process
option(BUILD_GO "Build qpid-proton-go shim" ON)

message(STATUS "BUILD_GO = ${BUILD_GO}")

set(gopath "${CMAKE_CURRENT_BINARY_DIR}")
# Stage the build

# CC=clang must be set for both cmake and make
# enable sanitizers; it slows the test down quite a lot
#set(FLAGS "-race")
#if (CMAKE_C_COMPILER_ID MATCHES "Clang")
#  # this requires building everything with -fsanitize=memory, that means special Dockerfile for me ...
#  set(FLAGS "-msan")
#endif (CMAKE_C_COMPILER_ID MATCHES "Clang")

macro(register_go_shim test binary)
  execute_process(
    COMMAND mkdir -p ${test}
#    COMMAND ln -sf "${CMAKE_CURRENT_SOURCE_DIR}/src/amqp_types_test" src/amqp_types_test
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
  )
  add_custom_target(
    go_${test}_${binary}
    ALL
    COMMAND ${CMAKE_COMMAND} -E env "GOPATH=${gopath}" go build ${FLAGS} ${ARGN} >> go_get.log
    DEPENDS
      ${ARGN}
      go_get_dependencies
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}/${test}
    COMMENT Compiling go ${test}/${binary}
  )

 install(
    FILES ${CMAKE_CURRENT_BINARY_DIR}/${test}/${binary}
    DESTINATION ${GO_INSTALL_ROOT}/amqp_types_test/
    PERMISSIONS WORLD_EXECUTE)
endmacro(register_go_shim)

if (BUILD_GO)
  # Set GOPATH to ${CMAKE_CURRENT_BINARY_DIR} and link the src/amqp_types_test subtree there

  # TODO: use PROTON_INSTALL_DIR in CFLAGS or somewhere to link with nonstandard location
  # Emit the custom build commands
  add_custom_target(
    go_get_dependencies
    ALL
    # http://stackoverflow.com/questions/35029277/how-to-modify-environment-variables-passed-to-custom-cmake-target
    COMMAND ${CMAKE_COMMAND} -E env "GOPATH=${gopath}" go get qpid.apache.org/electron > go_get.log
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
    COMMENT Fetching go dependencies
  )

 # Install
 set(GO_INSTALL_ROOT ${CMAKE_INSTALL_PREFIX}/libexec/qpid_interop_test/shims/qpid-proton-go)

 register_go_shim(amqp_types_test Sender ${CMAKE_CURRENT_SOURCE_DIR}/amqp_types_test/Sender.go ${CMAKE_CURRENT_SOURCE_DIR}/amqp_types_test/shim.go)
 register_go_shim(amqp_types_test Receiver ${CMAKE_CURRENT_SOURCE_DIR}/amqp_types_test/Receiver.go ${CMAKE_CURRENT_SOURCE_DIR}/amqp_types_test/shim.go)


endif ()
