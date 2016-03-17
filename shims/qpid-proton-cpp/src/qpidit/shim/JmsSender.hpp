/*
 *
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
 *
 */

#ifndef SRC_QPIDIT_SHIM_JMSSENDER_HPP_
#define SRC_QPIDIT_SHIM_JMSSENDER_HPP_

#include "json/value.h"
#include "proton/handler.hpp"
#include "qpidit/QpidItErrors.hpp"
#include <typeinfo>

namespace proton {
    class message;
}

namespace qpidit
{
    namespace shim
    {

        class JmsSender : public proton::handler
        {
        protected:
            static proton::symbol s_jmsMessageTypeAnnotationKey;
            static std::map<std::string, int8_t>s_jmsMessageTypeAnnotationValues;

            const std::string _brokerUrl;
            const std::string _jmsMessageType;
            const Json::Value _testValueMap;
            uint32_t _msgsSent;
            uint32_t _msgsConfirmed;
            uint32_t _totalMsgs;
        public:
            JmsSender(const std::string& brokerUrl, const std::string& jmsMessageType, const Json::Value& testValues);
            virtual ~JmsSender();
            void on_start(proton::event &e);
            void on_sendable(proton::event &e);
            void on_delivery_accept(proton::event &e);
            void on_disconnect(proton::event &e);
        protected:
            void  sendMessages(proton::event &e, const std::string& subType, const Json::Value& testValueMap);
            proton::message& setBytesMessage(proton::message& msg, const std::string& subType, const std::string& testValueStr);
            proton::message& setMapMessage(proton::message& msg, const std::string& subType, const std::string& testValueStr, uint32_t valueNumber);
            proton::message& setObjectMessage(proton::message& msg, const std::string& subType, const Json::Value& testValue);
            proton::message& setStreamMessage(proton::message& msg, const std::string& subType, const std::string& testValue);
            proton::message& setTextMessage(proton::message& msg, const Json::Value& testValue);

            static proton::binary getJavaObjectBinary(const std::string& javaClassName, const std::string& valAsString);
            static uint32_t getTotalNumMessages(const Json::Value& testValueMap);

            static std::map<std::string, int8_t> initializeJmsMessageTypeAnnotationMap();

            // Set message body to floating type T through integral type U
            // Used to convert a hex string representation of a float or double to a float or double
            template<typename T, typename U> T getFloatValue(const std::string& testValueStr) {
                try {
                    U ival(std::strtoul(testValueStr.data(), NULL, 16));
                    return T(*reinterpret_cast<T*>(&ival));
                } catch (const std::exception& e) { throw qpidit::InvalidTestValueError(typeid(T).name(), testValueStr); }
            }

            template<typename T> T getIntegralValue(const std::string& testValueStr) {
                try {
                    return T(std::strtol(testValueStr.data(), NULL, 16));
                } catch (const std::exception& e) { throw qpidit::InvalidTestValueError(typeid(T).name(), testValueStr); }
            }
        };

    } /* namespace shim */
} /* namespace qpidit */

#endif /* SRC_QPIDIT_SHIM_JMSSENDER_HPP_ */
