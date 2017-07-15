package corduroy;

import java.io.Serializable;
import java.lang.reflect.Type;

/**
 * Message represents information that is passed from {@link Node} to Node.
 */
public class Message implements Serializable {

    /**
     * Sets the payload that the message wraps.
     * @param o The payload.
     */
    public void setPayload(Serializable o) {
        payload = o;
        type = o.getClass();
    }

    /**
     * Gets the payload that the message wraps.
     * @param <T> The type to cast the payload into.
     * @return The payload.
     */
    public <T extends Serializable> T getPayload() {
        return (T) payload;
    }

    private Type type;

    private Object payload;
}
