// Copyright (c) 2023 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deltatracker

// SetDeltaTracker (conceptually) tracks the differences between two sets
// the "desired" set contains the members that we _want_ to be in the
// dataplane; the "dataplane" set contains the members that we think are
// _actually_ in the dataplane. The name "dataplane" set is intended to hint at
// its use but(!) this is a pure in-memory datastructure; it doesn't actually
// interact with the dataplane directly.
//
// The desired and dataplane sets are exposed via the Desired() and Dataplane()
// methods, which each return a similar set API featuring Set(...) Contains(...)
// Delete(...) and Iter(...). The dataplane set view has an additional
// ReplaceAllIter method, which allows for the whole contents of the
// dataplane set to be replaced via an iterator; this is more efficient than
// doing an external iteration and Set/Delete calls.
//
// In addition to the desired and dataplane sets, the differences between them
// are continuously tracked in two other sets: the "pending updates" set and
// the "pending deletions" set. "Pending updates" contains all keys that are
// in the "desired" set but not in the dataplane set. "Pending deletions" contains
// keys that are in the dataplane set but not in the desired set.  The
// pending sets are exposed via the IterPendingUpdates and IterPendingDeletions
// methods.
type SetDeltaTracker[K comparable] DeltaTracker[K, struct{}]

func NewSetDeltaTracker[K comparable]() *SetDeltaTracker[K] {
	// Implementation: we use a map delta tracker with empty struct as keys.
	dt := New[K, struct{}](WithValuesEqualFn[K, struct{}](func(a, b struct{}) bool {
		return true // empty struct always equals itself.
	}))
	return (*SetDeltaTracker[K])(dt)
}

type DesiredSetView[K comparable] DesiredView[K, struct{}]

func (s *SetDeltaTracker[K]) Desired() *DesiredSetView[K] {
	mapDT := (*DeltaTracker[K, struct{}])(s)
	return (*DesiredSetView[K])(mapDT.Desired())
}

func (s *DesiredSetView[K]) Add(k K) {
	s.asMapView().Set(k, struct{}{})
}

func (s *DesiredSetView[K]) Contains(k K) bool {
	_, exists := s.asMapView().Get(k)
	return exists
}

func (s *DesiredSetView[K]) Delete(k K) {
	s.asMapView().Delete(k)
}

func (s *DesiredSetView[K]) DeleteAll() {
	s.asMapView().DeleteAll()
}

func (s *DesiredSetView[K]) Iter(f func(k K)) {
	s.asMapView().Iter(func(k K, _ struct{}) {
		f(k)
	})
}

func (s *DesiredSetView[K]) asMapView() *DesiredView[K, struct{}] {
	return (*DesiredView[K, struct{}])(s)
}

type DataplaneSetView[K comparable] DesiredView[K, struct{}]

func (s *SetDeltaTracker[K]) Dataplane() *DataplaneSetView[K] {
	mapDT := (*DeltaTracker[K, struct{}])(s)
	return (*DataplaneSetView[K])(mapDT.Dataplane())
}

func (s *DataplaneSetView[K]) ReplaceFromIter(iter func(func(k K)) error) error {
	return s.asMapView().ReplaceAllIter(func(f func(k K, v struct{})) error {
		return iter(func(k K) {
			f(k, struct{}{})
		})
	})
}

func (s *DataplaneSetView[K]) Add(k K) {
	s.asMapView().Set(k, struct{}{})
}

func (s *DataplaneSetView[K]) Delete(k K) {
	s.asMapView().Delete(k)
}

func (s *DataplaneSetView[K]) Contains(k K) bool {
	_, exists := s.asMapView().Get(k)
	return exists
}

func (s *DataplaneSetView[K]) Iter(f func(k K)) {
	s.asMapView().Iter(func(k K, v struct{}) {
		f(k)
	})
}

func (s *DataplaneSetView[K]) asMapView() *DataplaneView[K, struct{}] {
	return (*DataplaneView[K, struct{}])(s)
}

func (s *SetDeltaTracker[K]) IterPendingUpdates(f func(k K) IterAction) {
	mapDT := (*DeltaTracker[K, struct{}])(s)
	mapDT.IterPendingUpdates(func(k K, v struct{}) IterAction {
		return f(k)
	})
}

func (s *SetDeltaTracker[K]) IterPendingDeletions(f func(k K) IterAction) {
	mapDT := (*DeltaTracker[K, struct{}])(s)
	mapDT.IterPendingDeletions(f)
}
