/*
 * Copyright 2020 Tero Vierimaa
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package modal contains modal views that are drawn on top of existing layout.
package modal

import "gitlab.com/tslocum/cview"

//Modal creates a modal that overlaps other views and get's destroyed when it's ready
type Modal interface {
	SetDoneFunc(doneFunc func())
	cview.Primitive
	View() cview.Primitive
}
