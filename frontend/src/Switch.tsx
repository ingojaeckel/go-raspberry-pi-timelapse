import React from 'react';
import { Switch, Route } from 'react-router-dom'
import PhotoComponent from './components/PhotoComponent';
import MonitoringComponent from './components/MonitoringComponent';
import SetupComponent from './components/SetupComponent';
import PreviewComponent from './components/PreviewComponent';
import LogComponent from './components/LogComponent';

export default function Switcher() {
    return (
        <Switch>
            <Route exact path={`${process.env.PUBLIC_URL}/`}>
                <PhotoComponent />
            </Route>
            <Route exact path={`${process.env.PUBLIC_URL}/preview`}>
                <PreviewComponent />
            </Route>
            <Route exact path={`${process.env.PUBLIC_URL}/monitoring`}>
                <MonitoringComponent />
            </Route>
            <Route exact path={`${process.env.PUBLIC_URL}/settings`}>
                <SetupComponent />
            </Route>
            <Route exact path={`${process.env.PUBLIC_URL}/logs`}>
                <LogComponent />
            </Route>
        </Switch>
    );
}
