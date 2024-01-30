import React from "react";

import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

import EasyDeploySvg from "@site/static/img/easy_deploy.svg";
import UniversalSupportSvg from "@site/static/img/universal_support.svg";
import FocusSvg from "@site/static/img/focus.svg";

type FeatureItem = {
  title: string;
  Svg: React.ComponentType<React.ComponentProps<'svg'>>;
  description: JSX.Element;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Easy to Use',
    Svg: EasyDeploySvg,
    description: (
      <>
        Create an interchain account (ICA) with a single instantiate call. No contracts are needed
        on the counterparty chain. Send ICA transactions as `CosmosMsg`s and receive callbacks.
      </>
    ),
  },
  {
    title: 'Universal CosmWasm Support',
    Svg: UniversalSupportSvg,
    description: (
      <>
        CosmWasm ICA Controller can be deployed on all IBC enabled CosmWasm chains. There is no need
        for custom chain bindings or even the interchain accounts module.
      </>
    ),
  },
  {
    title: 'Focus on What Matters',
    Svg: FocusSvg,
    description: (
      <>
        CosmWasm ICA Controller lets you focus on your application, and we&apos;ll do the IBC chores
        in the background. Go ahead and build your cross-chain application.
      </>
    ),
  },
];

function Feature({title, Svg, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): JSX.Element {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
