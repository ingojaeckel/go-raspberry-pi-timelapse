import React from 'react';
import { Switch, Route } from 'react-router-dom'
import PhotoComponent from './components/PhotoComponent';
import MonitoringComponent from './components/MonitoringComponent';
import SetupComponent from './components/SetupComponent';

export default function Switcher() {
    return (
        <Switch>
            <Route exact path="/">
                <PhotoComponent />
            </Route>
            <Route exact path="/monitoring">
                <MonitoringComponent />
            </Route>
            <Route exact path="/settings">
                <SetupComponent />
            </Route>
        </Switch>
    );
}
