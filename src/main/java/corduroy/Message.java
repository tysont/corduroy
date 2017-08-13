package corduroy;

import lombok.Getter;

import java.io.Serializable;
import java.lang.reflect.Type;

/**
 * Message represents information that is passed from {@link Node} to Node.
 */
public class Message implements Serializable {

    /**
     * Creates a message by providing a payload.
     * @param payload The payload.
     */
    public Message(Serializable payload) {
        this.payload = payload;
        type = payload.getClass();
    }

    /**
     * Gets the payload that the requestMessage wraps.
     * @param <T> The type to cast the payload into.
     * @return The payload.
     */
    public <T extends Serializable> T getPayload() {
        return (T) payload;
    }

    /**
     * The type of the message payload.
     */
    @Getter
    private Type type;

    /**
     * The message payload.
     */
    private Object payload;
}
