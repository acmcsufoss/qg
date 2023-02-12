export default {
  definitions: {
    Error: {
      additionalProperties: false,
      metadata: {
        description: "Error is returned on every API error.\n",
      },
      optionalProperties: {},
      properties: {
        message: {
          metadata: {
            description: "Message is the error message",
          },
          type: "string",
        },
      },
    },
    Game: {
      discriminator: "game",
      mapping: {
        jeopardy: {
          properties: {
            data: {
              ref: "Jeopardy",
            },
          },
        },
        kahoot: {
          properties: {
            data: {
              ref: "Kahoot",
            },
          },
        },
      },
      metadata: {
        description:
          "Game is the main game object. It contains all the information about the\ngame.\n",
      },
    },
    Jeopardy: {
      additionalProperties: false,
      metadata: {
        description: "JeopardyGame is the game data for a Jeopardy game.\n",
      },
      optionalProperties: {
        moderators: {
          metadata: {
            description: "moderators enables moderators being able to join.\n",
          },
          type: "boolean",
        },
        require_name: {
          metadata: {
            description:
              "require_name, if true, will require members to input a name before\nwe can participate.\n",
          },
          type: "boolean",
        },
        score_multiplier: {
          metadata: {
            description:
              "score_multiplier is the score multiplier for each question. The\ndefault is 100.\n",
          },
          type: "float32",
        },
      },
      properties: {
        categories: {
          elements: {
            ref: "JeopardyCategory",
          },
        },
      },
    },
    JeopardyCategory: {
      additionalProperties: false,
      metadata: {
        description: "JeopardyCategory is a category in a Jeopardy game.\n",
      },
      optionalProperties: {},
      properties: {
        name: {
          metadata: {
            description: "name is the name of the category.\n",
          },
          type: "string",
        },
        questions: {
          elements: {
            ref: "JeopardyQuestion",
          },
          metadata: {
            description: "questions are the questions in the category.\n",
          },
        },
      },
    },
    JeopardyQuestion: {
      additionalProperties: false,
      metadata: {
        description: "JeopardyQuestion is a question in a Jeopardy game.\n",
      },
      optionalProperties: {},
      properties: {
        answers: {
          elements: {
            type: "string",
          },
          metadata: {
            description: "answers are the possible answers.\n",
          },
        },
        correct_answer: {
          metadata: {
            description:
              "correct_answer is the correct answer within the list of answers\nabove. The index starts at 1.\n",
          },
          type: "int32",
        },
        question: {
          metadata: {
            description: "question is the question.\n",
          },
          type: "string",
        },
      },
    },
    Kahoot: {
      additionalProperties: false,
      metadata: {
        description: "KahootGame is the game data for a Kahoot game.\n",
      },
      optionalProperties: {},
      properties: {
        questions: {
          elements: {
            ref: "KahootQuestion",
          },
          metadata: {
            description: "questions are the questions in the game.\n",
          },
        },
        time_limit: {
          metadata: {
            description:
              "time_limit is the time limit for each question. The format is in\nGo's time.Duration, e.g. 10s for 10 seconds.\n",
          },
          type: "string",
        },
      },
    },
    KahootQuestion: {
      additionalProperties: false,
      metadata: {
        description: "KahootQuestion is a question in a Kahoot game.\n",
      },
      optionalProperties: {},
      properties: {
        answers: {
          elements: {
            type: "string",
          },
          metadata: {
            description: "answers are the possible answers.\n",
          },
        },
        question: {
          metadata: {
            description: "question is the question.\n",
          },
          type: "string",
        },
      },
    },
  },
};
