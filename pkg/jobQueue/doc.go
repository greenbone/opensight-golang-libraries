// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package jobQueue provides a thread-safe queue of requests to execute a predefined function.
// When a request is added to an empty queue, it is processed immediately.
// If there is already a request running, the new request will be executed after the current one.
// If several requests are waiting, only the last one is processed
package jobQueue
