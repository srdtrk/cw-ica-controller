import React from "react";
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "./ui/hover-card"

const tags = {
  concepts: {
    color: "#54ffe0",
    label: "Concepts",
    isBright: true,
    description: "Learn about the concepts behind 'cw-ica-controller'",
  },
  basics: {
    color: "#F69900",
    label: "Basics",
    isBright: true,
    description: "Learn the basics of 'cw-ica-controller'",
  },
  "ibc-go": {
    color: "#ff1717",
    label: "IBC-Go",
    description: "This section includes IBC-Go specific content",
  },
  cosmjs: {
    color: "#6836D0",
    label: "CosmJS",
    description: "This section includes CosmJS specific content",
  },
  cosmwasm: {
    color: "#05BDFC",
    label: "CosmWasm",
    description: "This section includes CosmWasm specific content",
  },
  protocol: {
    color: "#00B067",
    label: "Protocol",
    description: "This section includes content about protocol specifcations",
  },
  advanced: {
    color: "#f7f199",
    label: "Advanced",
    isBright: true,
    description: "The content in this section is for advanced users researching",
  },
  developer: {
    color: "#AABAFF",
    label: "Developer",
    isBright: true,
    description: "This section includes content for external developers using the 'cw-ica-controller'",
  },
  tutorial: {
    color: "#F46800",
    label: "Tutorial",
    description: "This section includes a tutorial",
  },
  "guided-coding": {
    color: "#F24CF4",
    label: "Guided Coding",
    description: "This section includes guided coding",
  },
};

const HighlightTag = ({ type, version }) => {
  const styles = tags[type] || tags["ibc-go"]; // default to 'ibc-go' if type doesn't exist
  const description = styles.description || "";

  return (
  <HoverCard>
    <HoverCardTrigger>
      <span
        style={{
          backgroundColor: styles.color,
          borderRadius: "2px",
          color: styles.isBright ? "black" : "white",
          padding: "0.3rem",
          marginBottom: "1rem",
          marginRight: "0.25rem",
          display: "inline-block",
        }}
      >
        {styles.label}
        {version ? ` ${version}` : ""}
      </span>
    </HoverCardTrigger>
    <HoverCardContent>
      {description}
    </HoverCardContent>
  </HoverCard>
  );
};

export default HighlightTag;
